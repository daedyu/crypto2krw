import Foundation

// MARK: - API Response

struct APIResponse<T: Decodable>: Decodable {
    let success: Bool
    let data: T?
    let error: APIError?
}

struct APIError: Decodable {
    let code: String?
    let message: String
}

private struct ErrorEnvelope: Decodable {
    struct Err: Decodable { let message: String }
    let error: Err?
}

// MARK: - API Client Errors

enum APIClientError: LocalizedError {
    case serverError(String)
    case networkError(Error)
    case decodingError(Error)
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .serverError(let msg): return msg
        case .networkError(let e): return e.localizedDescription
        case .decodingError(let e): return "응답 파싱 오류: \(e.localizedDescription)"
        case .unauthorized: return "인증이 만료되었습니다. 다시 로그인해주세요."
        }
    }
}

// MARK: - Token Storage

enum TokenStore {
    private static let accessKey  = "crypto2krw.access_token"
    private static let refreshKey = "crypto2krw.refresh_token"

    static var accessToken: String? {
        get { UserDefaults.standard.string(forKey: accessKey) }
        set { UserDefaults.standard.set(newValue, forKey: accessKey) }
    }

    static var refreshToken: String? {
        get { UserDefaults.standard.string(forKey: refreshKey) }
        set { UserDefaults.standard.set(newValue, forKey: refreshKey) }
    }

    static func clear() {
        UserDefaults.standard.removeObject(forKey: accessKey)
        UserDefaults.standard.removeObject(forKey: refreshKey)
    }
}

// MARK: - API Client

@MainActor
final class APIClient {
    static let shared = APIClient()

    private let baseURL: String
    private let session: URLSession
    private let decoder: JSONDecoder

    private init() {
        baseURL = ProcessInfo.processInfo.environment["API_BASE_URL"] ?? "http://192.168.0.18:3000"
        session = .shared
        decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        decoder.keyDecodingStrategy = .convertFromSnakeCase
    }

    // MARK: - Core Request

    func request<T: Decodable>(
        _ path: String,
        method: String = "GET",
        body: Encodable? = nil,
        requiresAuth: Bool = true
    ) async throws -> T {
        guard let url = URL(string: baseURL + path) else {
            throw APIClientError.serverError("잘못된 URL")
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if requiresAuth, let token = TokenStore.accessToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body {
            let encoder = JSONEncoder()
            encoder.keyEncodingStrategy = .convertToSnakeCase
            urlRequest.httpBody = try encoder.encode(body)
        }

        do {
            let (data, response) = try await session.data(for: urlRequest)
            guard let http = response as? HTTPURLResponse else {
                throw APIClientError.serverError("잘못된 응답")
            }

            if http.statusCode == 401 {
                // 토큰 만료 시 자동 갱신 시도
                if requiresAuth, let _ = TokenStore.refreshToken {
                    try await refreshTokens()
                    return try await request(path, method: method, body: body, requiresAuth: requiresAuth)
                }
                // 서버가 보낸 실제 에러 메시지 우선 사용 (로그인 실패 등)
                if let envelope = try? decoder.decode(ErrorEnvelope.self, from: data),
                   let msg = envelope.error?.message {
                    throw APIClientError.serverError(msg)
                }
                throw APIClientError.unauthorized
            }

            let apiResp = try decoder.decode(APIResponse<T>.self, from: data)
            if let data = apiResp.data {
                return data
            }
            throw APIClientError.serverError(apiResp.error?.message ?? "알 수 없는 오류")
        } catch let error as APIClientError {
            throw error
        } catch let error as DecodingError {
            throw APIClientError.decodingError(error)
        } catch {
            throw APIClientError.networkError(error)
        }
    }

    // MARK: - Token Refresh

    private func refreshTokens() async throws {
        guard let refresh = TokenStore.refreshToken else {
            throw APIClientError.unauthorized
        }

        struct RefreshBody: Encodable { let refreshToken: String }
        struct TokenPair: Decodable { let accessToken: String; let refreshToken: String }

        let tokens: TokenPair = try await request(
            "/api/v1/auth/refresh",
            method: "POST",
            body: RefreshBody(refreshToken: refresh),
            requiresAuth: false
        )
        TokenStore.accessToken = tokens.accessToken
        TokenStore.refreshToken = tokens.refreshToken
    }
}
