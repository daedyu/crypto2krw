import SwiftUI

struct HistoryView: View {
    private let transactions = MockData.transactions

    var body: some View {
        List {
            ForEach(transactions) { tx in
                TransactionRow(tx: tx)
            }
        }
        .navigationTitle("결제 내역")
    }
}

#Preview {
    NavigationStack {
        HistoryView()
    }
}
