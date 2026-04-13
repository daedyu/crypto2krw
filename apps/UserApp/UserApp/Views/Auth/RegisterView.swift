import SwiftUI

struct RegisterView: View {
    @Environment(AuthManager.self) private var auth

    @State private var email           = ""
    @State private var password        = ""
    @State private var passwordConfirm = ""

    private var mismatch: Bool {
        !passwordConfirm.isEmpty && password != passwordConfirm
    }
    private var canSubmit: Bool {
        !email.isEmpty && !password.isEmpty && !passwordConfirm.isEmpty && !mismatch
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

                SecureField("비밀번호 확인", text: $passwordConfirm)
                    .textContentType(.newPassword)

                if mismatch {
                    Text("비밀번호가 일치하지 않습니다")
                        .font(.caption)
                        .foregroundStyle(.red)
                }
            } header: {
                Text("회원가입")
                    .font(.largeTitle).bold()
                    .foregroundStyle(.primary)
                    .textCase(nil)
                    .padding(.bottom, 8)
            }

            Section {
                Button("가입하기") {
                    auth.login(email: email, password: password)
                }
                .buttonStyle(.borderedProminent)
                .disabled(!canSubmit)
                .frame(maxWidth: .infinity)
            }

            Section {
                Label("가입 즉시 SOL · USDT · ETH 입금 주소가 자동으로 발급됩니다", systemImage: "wallet.pass")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .navigationTitle("회원가입")
        .navigationBarTitleDisplayMode(.inline)
    }
}

#Preview {
    NavigationStack {
        RegisterView()
            .environment(AuthManager())
    }
}
