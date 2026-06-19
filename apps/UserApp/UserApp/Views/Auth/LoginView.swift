import SwiftUI

struct LoginView: View {
    @Environment(AuthManager.self) private var auth
    @State private var email    = ""
    @State private var password = ""
    @State private var showRegister = false

    private var canLogin: Bool { !email.isEmpty && password.count >= 6 && !auth.isLoading }

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

                if let errorMessage = auth.errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }

                Section {
                    Button {
                        Task { await auth.login(email: email, password: password) }
                    } label: {
                        if auth.isLoading {
                            ProgressView()
                                .frame(maxWidth: .infinity)
                        } else {
                            Text("로그인")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(!canLogin)

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
