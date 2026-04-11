'use strict';

class AxiomConductorClient {
    constructor(options) {
        this.baseUrl = (options.baseUrl || process.env.AXIOM_BASE_URL || 'http://localhost:8000').replace(/\/$/, '');
        this.username = options.username || process.env.AXIOM_USERNAME;
        this.password = options.password || process.env.AXIOM_PASSWORD;
        this.accessToken = null;
        this.refreshToken = null;
        this.tokenType = 'Bearer';
        this.expiresAtMs = 0;
    }

    async login() {
        if (!this.username || !this.password) {
            throw new Error('Missing AXIOM_USERNAME or AXIOM_PASSWORD');
        }

        const res = await this._request('POST', '/auth/login', {
            username: this.username,
            password: this.password,
        }, false);

        this.accessToken = res.access_token;
        this.refreshToken = res.refresh_token || null;
        this.tokenType = res.token_type || 'Bearer';
        const expiresIn = Number(res.expires_in || 0);
        this.expiresAtMs = Date.now() + Math.max(expiresIn - 20, 20) * 1000;

        return res;
    }

    async refresh() {
        if (!this.refreshToken) {
            return this.login();
        }

        const res = await this._request('POST', '/auth/refresh', {
            refresh_token: this.refreshToken,
        }, false);

        this.accessToken = res.access_token;
        this.refreshToken = res.refresh_token || this.refreshToken;
        this.tokenType = res.token_type || 'Bearer';
        const expiresIn = Number(res.expires_in || 0);
        this.expiresAtMs = Date.now() + Math.max(expiresIn - 20, 20) * 1000;

        return res;
    }

    async ensureToken() {
        if (!this.accessToken) {
            await this.login();
            return;
        }
        if (Date.now() >= this.expiresAtMs) {
            await this.refresh();
        }
    }

    async getStats() {
        return this._request('GET', '/api/v1/conductor/stats');
    }

    async listProducers() {
        return this._request('GET', '/api/v1/conductor/producers');
    }

    async createProducer(payload) {
        return this._request('POST', '/api/v1/conductor/producers', payload);
    }

    async listConsumers() {
        return this._request('GET', '/api/v1/conductor/consumers');
    }

    async createConsumer(payload) {
        return this._request('POST', '/api/v1/conductor/consumers', payload);
    }

    async publish(payload) {
        return this._request('POST', '/api/v1/conductor/publish', payload);
    }

    async connectRabbitMQ(url) {
        return this._request('POST', '/api/v1/conductor/connections', {
            type: 'rabbitmq',
            url,
        });
    }

    async connectKafka(brokers) {
        return this._request('POST', '/api/v1/conductor/connections', {
            type: 'kafka',
            brokers,
        });
    }

    async listMessages(limit = 100) {
        return this._request('GET', '/api/v1/conductor/messages?limit=' + encodeURIComponent(String(limit)));
    }

    async listDLQ() {
        return this._request('GET', '/api/v1/conductor/dlq');
    }

    async replayDLQ(dlqId) {
        return this._request('POST', '/api/v1/conductor/dlq/' + encodeURIComponent(dlqId) + '/replay');
    }

    async getWebSocketStreamUrl() {
        await this.ensureToken();
        const wsBase = this.baseUrl.replace(/^http/i, 'ws');
        return wsBase + '/ws/conductor?token=' + encodeURIComponent(this.accessToken);
    }

    async _request(method, path, body, auth = true) {
        if (auth) {
            await this.ensureToken();
        }

        const headers = {
            'Content-Type': 'application/json',
        };

        if (auth && this.accessToken) {
            headers.Authorization = (this.tokenType || 'Bearer') + ' ' + this.accessToken;
        }

        const response = await fetch(this.baseUrl + path, {
            method,
            headers,
            body: body === undefined ? undefined : JSON.stringify(body),
        });

        const text = await response.text();
        let json;
        try {
            json = text ? JSON.parse(text) : {};
        } catch (_err) {
            json = { raw: text };
        }

        if (!response.ok) {
            const message = (json && (json.error || json.message)) || response.statusText || 'Request failed';
            throw new Error(method + ' ' + path + ' failed: ' + message);
        }

        return json;
    }
}

async function demo() {
    const client = new AxiomConductorClient({});

    const loginRes = await client.login();
    console.log('Logged in as:', loginRes.username || 'unknown', 'role:', loginRes.role || 'unknown');

    const stats = await client.getStats();
    console.log('Current Conductor stats:', stats);

    const producerId = process.env.AXIOM_DEMO_PRODUCER_ID;
    if (producerId) {
        const msg = await client.publish({
            producerId,
            body: {
                source: 'nodejs-client',
                ts: new Date().toISOString(),
                message: 'hello from Node.js client',
            },
            headers: {
                app: 'nodejs-demo',
            },
            correlationId: 'node-' + Date.now(),
        });
        console.log('Published message:', msg.id || msg);
    } else {
        console.log('Set AXIOM_DEMO_PRODUCER_ID to publish a test event.');
    }

    const wsUrl = await client.getWebSocketStreamUrl();
    console.log('WebSocket stream URL:', wsUrl);
}

if (require.main === module) {
    demo().catch((err) => {
        console.error(err.message || err);
        process.exit(1);
    });
}

module.exports = {
    AxiomConductorClient,
};
