import SwiftUI

// MARK: - Auth DTOs

private struct RegisterRequest: Encodable {
    let email: String
    let password: String
}

private struct LoginRequest: Encodable {
    let email: String
    let password: String
}

private struct AuthTokens: Decodable {
    let accessToken: String
    let refreshToken: String
    let user: UserInfo?

    struct UserInfo: Decodable {
        let id: String
        let email: String
    }
}

// MARK: - AuthManager

@MainActor
@Observable
final class AuthManager {
    var isLoggedIn: Bool = false
    var userEmail: String = ""
    var isLoading: Bool = false
    var errorMessage: String? = nil

    init() {
        // 저장된 토큰이 있으면 자동 로그인 상태로 복원
        if TokenStore.accessToken != nil {
            isLoggedIn = true
        }
    }

    func login(email: String, password: String) async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let tokens: AuthTokens = try await APIClient.shared.request(
                "/api/v1/auth/login",
                method: "POST",
                body: LoginRequest(email: email, password: password),
                requiresAuth: false
            )
            TokenStore.accessToken  = tokens.accessToken
            TokenStore.refreshToken = tokens.refreshToken
            userEmail  = email
            isLoggedIn = true
        } catch {
            errorMessage = (error as? APIClientError)?.errorDescription ?? error.localizedDescription
        }
    }

    func register(email: String, password: String) async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let tokens: AuthTokens = try await APIClient.shared.request(
                "/api/v1/auth/register",
                method: "POST",
                body: RegisterRequest(email: email, password: password),
                requiresAuth: false
            )
            TokenStore.accessToken  = tokens.accessToken
            TokenStore.refreshToken = tokens.refreshToken
            userEmail  = email
            isLoggedIn = true
        } catch {
            errorMessage = (error as? APIClientError)?.errorDescription ?? error.localizedDescription
        }
    }

    func logout() async {
        // 서버 측 refresh token 삭제 (실패해도 로컬에서 로그아웃)
        _ = try? await APIClient.shared.request(
            "/api/v1/auth/logout",
            method: "POST",
            body: Optional<String>.none
        ) as EmptyResponse

        TokenStore.clear()
        userEmail  = ""
        isLoggedIn = false
    }
}

// 빈 응답을 처리하기 위한 더미 타입
struct EmptyResponse: Decodable {}
