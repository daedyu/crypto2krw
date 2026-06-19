import SwiftUI

enum AppTab: Hashable {
    case home, history, settings, deposit
}

struct RootView: View {
    @Environment(AuthManager.self) private var auth

    var body: some View {
        if auth.isLoggedIn {
            MainTabView()
        } else {
            LoginView()
        }
    }
}

struct MainTabView: View {
    @State private var selectedTab: AppTab = .home
    @State private var showDeposit = false

    var body: some View {
        TabView(selection: $selectedTab) {
            Tab("결제", systemImage: "creditcard.fill", value: AppTab.home) {
                NavigationStack { HomeView() }
            }
            Tab("내역", systemImage: "list.bullet.rectangle", value: AppTab.history) {
                NavigationStack { HistoryView() }
            }
            Tab("설정", systemImage: "gearshape", value: AppTab.settings) {
                NavigationStack { SettingsView() }
            }
            Tab("입금", systemImage: "plus", value: AppTab.deposit, role: TabRole.search) {
                Color.orange.ignoresSafeArea()
            }
        }
        .onChange(of: selectedTab) { _, newValue in
            if newValue == .deposit {
                showDeposit = true
                selectedTab = .home
            }
        }
        .sheet(isPresented: $showDeposit) {
            NavigationStack { DepositListView() }
        }
    }
}

#Preview {
    MainTabView()
}

#Preview {
    MainTabView()
}
