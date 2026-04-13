import SwiftUI

struct LoginView: View {
    @Environment(AuthManager.self) private var auth
    @State private var email    = ""
    @State private var password = ""
    @State private var showRegister = false

    private var canLogin: Bool { !email.isEmpty && !password.isEmpty }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("이메일", text: $email)
                        .keyboardType(.emailAddress)
                        .textContentType(.emailAddress)
                        .autocorrectionDisabled()
                        .textInputAutocapitalization(.never)

                    SecureField("비밀번호", text: $password)
                        .textContentType(.password)
                } header: {
                    Text("로그인")
                        .font(.largeTitle).bold()
                        .foregroundStyle(.primary)
                        .textCase(nil)
                        .padding(.bottom, 8)
                }

                Section {
                    Button("로그인") {
                        auth.login(email: email, password: password)
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(!canLogin)
                    .frame(maxWidth: .infinity)

                    Button("계정이 없으신가요? 회원가입") {
                        showRegister = true
                    }
                    .frame(maxWidth: .infinity)
                }
            }
            .navigationBarHidden(true)
            .navigationDestination(isPresented: $showRegister) {
                RegisterView()
            }
        }
    }
}

#Preview {
    LoginView()
        .environment(AuthManager())
}
