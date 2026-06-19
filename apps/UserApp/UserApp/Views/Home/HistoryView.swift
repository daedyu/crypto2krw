import SwiftUI

struct HistoryView: View {
    @State private var transactions: [Transaction] = []
    @State private var isLoading = false

    var body: some View {
        List {
            if isLoading && transactions.isEmpty {
                HStack {
                    Spacer()
                    ProgressView()
                    Spacer()
                }
                .listRowBackground(Color.clear)
            } else if transactions.isEmpty {
                Text("거래 내역이 없습니다")
                    .foregroundStyle(.secondary)
                    .font(.subheadline)
            } else {
                ForEach(transactions) { tx in
                    TransactionRow(tx: tx)
                }
            }
        }
        .navigationTitle("결제 내역")
        .refreshable { await fetchTransactions() }
        .task { await fetchTransactions() }
    }

    private func fetchTransactions() async {
        isLoading = true
        defer { isLoading = false }
        if let items = try? await UserService.shared.getTransactions(limit: 100) {
            transactions = items.compactMap { $0.toTransaction() }
        }
    }
}

#Preview {
    NavigationStack {
        HistoryView()
    }
}
