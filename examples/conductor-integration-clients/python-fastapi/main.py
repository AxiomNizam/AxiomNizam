import os
import time
from typing import Any, Dict, Optional

import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field


class AxiomConductorClient:
    def __init__(self, base_url: str, username: str, password: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None
        self.token_type: str = "Bearer"
        self.expires_at: float = 0.0
        self.http = httpx.AsyncClient(timeout=30.0)

    async def login(self) -> Dict[str, Any]:
        payload = {"username": self.username, "password": self.password}
        res = await self.http.post(f"{self.base_url}/auth/login", json=payload)
        data = self._parse_or_raise(res, "POST /auth/login")

        self.access_token = data.get("access_token")
        self.refresh_token = data.get("refresh_token")
        self.token_type = data.get("token_type") or "Bearer"
        expires_in = int(data.get("expires_in") or 0)
        self.expires_at = time.time() + max(expires_in - 20, 20)
        return data

    async def refresh(self) -> Dict[str, Any]:
        if not self.refresh_token:
            return await self.login()

        payload = {"refresh_token": self.refresh_token}
        res = await self.http.post(f"{self.base_url}/auth/refresh", json=payload)
        data = self._parse_or_raise(res, "POST /auth/refresh")

        self.access_token = data.get("access_token")
        self.refresh_token = data.get("refresh_token") or self.refresh_token
        self.token_type = data.get("token_type") or "Bearer"
        expires_in = int(data.get("expires_in") or 0)
        self.expires_at = time.time() + max(expires_in - 20, 20)
        return data

    async def ensure_token(self) -> None:
        if not self.access_token:
            await self.login()
            return
        if time.time() >= self.expires_at:
            await self.refresh()

    async def request(self, method: str, path: str, payload: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        await self.ensure_token()

        headers = {
            "Authorization": f"{self.token_type} {self.access_token}",
            "Content-Type": "application/json",
        }
        res = await self.http.request(method, f"{self.base_url}{path}", json=payload, headers=headers)

        if res.status_code == 401:
            # Token may be expired/revoked; refresh once and retry.
            await self.refresh()
            headers["Authorization"] = f"{self.token_type} {self.access_token}"
            res = await self.http.request(method, f"{self.base_url}{path}", json=payload, headers=headers)

        return self._parse_or_raise(res, f"{method} {path}")

    async def get_stats(self) -> Dict[str, Any]:
        return await self.request("GET", "/api/v1/conductor/stats")

    async def publish(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return await self.request("POST", "/api/v1/conductor/publish", payload)

    async def connect_rabbitmq(self, url: str) -> Dict[str, Any]:
        return await self.request("POST", "/api/v1/conductor/connections", {"type": "rabbitmq", "url": url})

    async def create_producer(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return await self.request("POST", "/api/v1/conductor/producers", payload)

    async def ws_stream_url(self) -> str:
        await self.ensure_token()
        ws_base = self.base_url.replace("https://", "wss://").replace("http://", "ws://")
        return f"{ws_base}/ws/conductor?token={self.access_token}"

    @staticmethod
    def _parse_or_raise(response: httpx.Response, op: str) -> Dict[str, Any]:
        try:
            data = response.json()
        except ValueError:
            data = {"raw": response.text}

        if response.status_code >= 400:
            msg = ""
            if isinstance(data, dict):
                msg = str(data.get("error") or data.get("message") or "")
            if not msg:
                msg = f"status={response.status_code}"
            raise HTTPException(status_code=response.status_code, detail=f"{op} failed: {msg}")

        if isinstance(data, dict):
            return data
        return {"data": data}


class PublishRequest(BaseModel):
    producer_id: str = Field(alias="producerId")
    body: Dict[str, Any]
    correlation_id: Optional[str] = Field(default=None, alias="correlationId")


class RabbitMQConnectRequest(BaseModel):
    url: str


AXIOM_BASE_URL = os.getenv("AXIOM_BASE_URL", "http://localhost:8000")
AXIOM_USERNAME = os.getenv("AXIOM_USERNAME", "")
AXIOM_PASSWORD = os.getenv("AXIOM_PASSWORD", "")

if not AXIOM_USERNAME or not AXIOM_PASSWORD:
    raise RuntimeError("Set AXIOM_USERNAME and AXIOM_PASSWORD before starting FastAPI app")

client = AxiomConductorClient(
    base_url=AXIOM_BASE_URL,
    username=AXIOM_USERNAME,
    password=AXIOM_PASSWORD,
)

app = FastAPI(title="AxiomNizam Conductor IAM Integration")


@app.on_event("startup")
async def startup() -> None:
    await client.login()


@app.on_event("shutdown")
async def shutdown() -> None:
    await client.http.aclose()


@app.get("/health")
async def health() -> Dict[str, str]:
    return {"status": "ok"}


@app.get("/conductor/stats")
async def conductor_stats() -> Dict[str, Any]:
    return await client.get_stats()


@app.post("/conductor/connections/rabbitmq")
async def conductor_connect_rabbitmq(req: RabbitMQConnectRequest) -> Dict[str, Any]:
    return await client.connect_rabbitmq(req.url)


@app.post("/conductor/producers")
async def conductor_create_producer(payload: Dict[str, Any]) -> Dict[str, Any]:
    return await client.create_producer(payload)


@app.post("/conductor/publish")
async def conductor_publish(req: PublishRequest) -> Dict[str, Any]:
    payload = {
        "producerId": req.producer_id,
        "body": req.body,
    }
    if req.correlation_id:
        payload["correlationId"] = req.correlation_id
    return await client.publish(payload)


@app.get("/conductor/stream/ws-url")
async def conductor_stream_ws_url() -> Dict[str, str]:
    return {"ws_url": await client.ws_stream_url()}
