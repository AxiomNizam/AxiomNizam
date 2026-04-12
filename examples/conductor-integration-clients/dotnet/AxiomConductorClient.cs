using System.Net;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace AxiomNizam.Conductor;

public sealed class AxiomConductorClient
{
    private readonly HttpClient _http;
    private readonly string _username;
    private readonly string _password;

    private string? _accessToken;
    private string? _refreshToken;
    private string _tokenType = "Bearer";
    private DateTimeOffset _expiresAt = DateTimeOffset.MinValue;

    public AxiomConductorClient(HttpClient httpClient, string username, string password)
    {
        _http = httpClient;
        _username = username;
        _password = password;

        if (_http.BaseAddress is null)
        {
            _http.BaseAddress = new Uri("http://localhost:8000");
        }
    }

    public async Task<JsonDocument> LoginAsync(CancellationToken cancellationToken = default)
    {
        var payload = new
        {
            username = _username,
            password = _password
        };

        var doc = await SendJsonAsync(HttpMethod.Post, "/auth/login", payload, auth: false, cancellationToken);
        UpdateTokenState(doc.RootElement);
        return doc;
    }

    public async Task<JsonDocument> RefreshAsync(CancellationToken cancellationToken = default)
    {
        if (string.IsNullOrWhiteSpace(_refreshToken))
        {
            return await LoginAsync(cancellationToken);
        }

        var payload = new
        {
            refresh_token = _refreshToken
        };

        var doc = await SendJsonAsync(HttpMethod.Post, "/auth/refresh", payload, auth: false, cancellationToken);
        UpdateTokenState(doc.RootElement);
        return doc;
    }

    public async Task<JsonDocument> GetStatsAsync(CancellationToken cancellationToken = default)
    {
        await EnsureTokenAsync(cancellationToken);
        return await SendJsonAsync(HttpMethod.Get, "/api/v1/conductor/stats", null, auth: true, cancellationToken);
    }

    public async Task<JsonDocument> ListProducersAsync(CancellationToken cancellationToken = default)
    {
        await EnsureTokenAsync(cancellationToken);
        return await SendJsonAsync(HttpMethod.Get, "/api/v1/conductor/producers", null, auth: true, cancellationToken);
    }

    public async Task<JsonDocument> CreateProducerAsync(object payload, CancellationToken cancellationToken = default)
    {
        await EnsureTokenAsync(cancellationToken);
        return await SendJsonAsync(HttpMethod.Post, "/api/v1/conductor/producers", payload, auth: true, cancellationToken);
    }

    public async Task<JsonDocument> PublishAsync(object payload, CancellationToken cancellationToken = default)
    {
        await EnsureTokenAsync(cancellationToken);
        return await SendJsonAsync(HttpMethod.Post, "/api/v1/conductor/publish", payload, auth: true, cancellationToken);
    }

    public async Task<JsonDocument> ConnectRabbitMqAsync(string url, CancellationToken cancellationToken = default)
    {
        await EnsureTokenAsync(cancellationToken);
        return await SendJsonAsync(
            HttpMethod.Post,
            "/api/v1/conductor/connections",
            new { type = "rabbitmq", url },
            auth: true,
            cancellationToken
        );
    }

    public async Task<string> GetWebSocketStreamUrlAsync(CancellationToken cancellationToken = default)
    {
        await EnsureTokenAsync(cancellationToken);
        var baseUrl = _http.BaseAddress!.ToString().TrimEnd('/');
        var wsBase = baseUrl.StartsWith("https://", StringComparison.OrdinalIgnoreCase)
            ? "wss://" + baseUrl[8..]
            : "ws://" + baseUrl.Replace("http://", "", StringComparison.OrdinalIgnoreCase);

        return $"{wsBase}/ws/conductor?token={Uri.EscapeDataString(_accessToken ?? string.Empty)}";
    }

    private async Task EnsureTokenAsync(CancellationToken cancellationToken)
    {
        if (string.IsNullOrWhiteSpace(_accessToken))
        {
            await LoginAsync(cancellationToken);
            return;
        }

        if (DateTimeOffset.UtcNow >= _expiresAt)
        {
            await RefreshAsync(cancellationToken);
        }
    }

    private void UpdateTokenState(JsonElement root)
    {
        _accessToken = root.TryGetProperty("access_token", out var at) ? at.GetString() : null;
        _refreshToken = root.TryGetProperty("refresh_token", out var rt) ? rt.GetString() : _refreshToken;
        _tokenType = root.TryGetProperty("token_type", out var tt) && !string.IsNullOrWhiteSpace(tt.GetString())
            ? tt.GetString()!
            : "Bearer";

        var expiresIn = root.TryGetProperty("expires_in", out var ei) && ei.TryGetInt32(out var v) ? v : 0;
        var safeTtl = Math.Max(expiresIn - 20, 20);
        _expiresAt = DateTimeOffset.UtcNow.AddSeconds(safeTtl);
    }

    private async Task<JsonDocument> SendJsonAsync(
        HttpMethod method,
        string path,
        object? payload,
        bool auth,
        CancellationToken cancellationToken,
        bool retryOnUnauthorized = true)
    {
        using var req = new HttpRequestMessage(method, path);

        if (auth && !string.IsNullOrWhiteSpace(_accessToken))
        {
            req.Headers.Authorization = new AuthenticationHeaderValue(_tokenType, _accessToken);
        }

        if (payload is not null)
        {
            req.Content = new StringContent(JsonSerializer.Serialize(payload), Encoding.UTF8, "application/json");
        }

        using var res = await _http.SendAsync(req, cancellationToken);
        var text = await res.Content.ReadAsStringAsync(cancellationToken);

        if (!res.IsSuccessStatusCode)
        {
            if (auth && retryOnUnauthorized && res.StatusCode == HttpStatusCode.Unauthorized)
            {
                await RefreshAsync(cancellationToken);
                return await SendJsonAsync(method, path, payload, auth: true, cancellationToken, retryOnUnauthorized: false);
            }

            throw new InvalidOperationException($"{method} {path} failed ({(int)res.StatusCode}): {text}");
        }

        return JsonDocument.Parse(string.IsNullOrWhiteSpace(text) ? "{}" : text);
    }
}

/*
Example usage in Program.cs:

var baseUrl = Environment.GetEnvironmentVariable("AXIOM_BASE_URL") ?? "http://localhost:8000";
var username = Environment.GetEnvironmentVariable("AXIOM_USERNAME") ?? throw new Exception("AXIOM_USERNAME missing");
var password = Environment.GetEnvironmentVariable("AXIOM_PASSWORD") ?? throw new Exception("AXIOM_PASSWORD missing");

var http = new HttpClient { BaseAddress = new Uri(baseUrl) };
var client = new AxiomConductorClient(http, username, password);
await client.LoginAsync();
var stats = await client.GetStatsAsync();
Console.WriteLine(stats.RootElement);
*/
