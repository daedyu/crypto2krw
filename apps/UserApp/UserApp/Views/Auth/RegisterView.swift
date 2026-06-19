import SwiftUI

struct RegisterView: View {
    @Environment(AuthManager.self) private var auth
    @Environment(\.dismiss) private var dismiss
    @State private var email     = ""
    @State private var password  = ""
    @State private var confirm   = ""

    private var passwordsMatch: Bool { password == confirm }
    private var canRegister: Bool {
        !email.isEmpty && password.count >= 8 && passwordsMatch && !auth.isLoading
    }

    var body: some View {
        Form {
            Section {
                TextField("이메일", text: $email)
                    .keyboardType(.emailAddress)
                    .textContentType(.emailAddress)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)

                SecureField("비밀번호 (8자 이상)", text: $password)
                    .textContentType(.newPassword)

                SecureField("비밀번호 확인", text: $confirm)
                    .textContentType(.newPassword)
            } header: {
                Text("회원가입")
                    .font(.largeTitle).bold()
                    .foregroundStyle(.primary)
                    .textCase(nil)
                    .padding(.bottom, 8)
            }

            if !confirm.isEmpty && !passwordsMatch {
                Section {
                    Text("비밀번호가 일치하지 않습니다.")
                        .foregroundStyle(.red)
                        .font(.caption)
                }
            }

            if let errorMessage = auth.errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                        .font(.caption)
                }
            }

            Section {
                Button {
                    Task { await auth.register(email: email, password: password) }
                } label: {
                    if auth.isLoading {
                        ProgressView().frame(maxWidth: .infinity)
                    } else {
                        Text("가입하기").frame(maxWidth: .infinity)
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(!canRegister)
            }

            Section {
                Label("가입 즉시 SOL · USDT · ETH 입금 주소가 자동으로 발급됩니다", systemImage: "wallet.pass")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .navigationTitle("회원가입")
        .navigationBarTitleDisplayMode(.inline)
        .onChange(of: auth.isLoggedIn) { _, isLoggedIn in
            if isLoggedIn { dismiss() }
        }
    }
}

#Preview {
    NavigationStack {
        RegisterView()
            .environment(AuthManager())
    }
}
