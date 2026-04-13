import SwiftUI

@Observable
final class AuthManager {
    var isLoggedIn: Bool = false
    var userEmail: String = ""

    func login(email: String, password: String) {
        userEmail = email
        isLoggedIn = true
    }

    func logout() {
        userEmail = ""
        isLoggedIn = false
    }
}
