import SwiftUI

// MARK: - Scroll offset preference key

private struct ScrollOffsetKey: PreferenceKey {
    static let defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

// MARK: - Card Detail View

struct CardDetailView: View {
    let balance:      CoinBalance
    let cardH:        CGFloat
    let detailY:      CGFloat
    let screenWidth:  CGFloat
    let transactions: [Transaction]
    let onTapCard:    () -> Void

    @State private var scrollOffset: CGFloat = 0

    private let balanceH: CGFloat = 80

    // 스크롤 0 → opacity 1, 스크롤 -60 → opacity 0
    private var balanceOpacity: Double {
        max(0, min(1, 1 + scrollOffset / 60))
    }

    // 스크롤에 따라 살짝 위로 올라가는 parallax
    private var balanceParallax: CGFloat {
        min(0, scrollOffset * 0.35)
    }

    var body: some View {
        ZStack(alignment: .top) {

            // ── 레이아웃: 카드 공간 확보 + 스크롤 영역 ──
            // ScrollView를 카드+잔액 영역 아래에 물리적으로 배치해
            // 오버스크롤 rubber-band가 카드 위로 올라오는 것을 원천 차단
            VStack(spacing: 0) {
                Color.clear
                    .frame(height: detailY + cardH + 28 + balanceH)

                ScrollView(.vertical, showsIndicators: false) {
                    VStack(spacing: 0) {
                        // 거래 내역
                        if transactions.isEmpty {
                            VStack(spacing: 8) {
                                Image(systemName: "tray")
                                    .font(.system(size: 32))
                                    .foregroundStyle(.tertiary)
                                Text("거래 내역이 없습니다")
                                    .font(.subheadline)
                                    .foregroundStyle(.secondary)
                            }
                            .padding(.top, 40)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(Array(transactions.enumerated()), id: \.element.id) { i, tx in
                                    TransactionRow(tx: tx)
                                        .padding(.horizontal, 16)
                                    if i < transactions.count - 1 {
                                        Divider().padding(.leading, 72)
                                    }
                                }
                            }
                            .background(Color(.systemBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
                            .padding(.horizontal, 20)
                        }

                        Spacer(minLength: 48)
                    }
                    // 스크롤 offset 추적
                    .background(
                        GeometryReader { proxy in
                            Color.clear.preference(
                                key: ScrollOffsetKey.self,
                                value: proxy.frame(in: .named("detailScroll")).minY
                            )
                        }
                    )
                }
                .coordinateSpace(name: "detailScroll")
                .onPreferenceChange(ScrollOffsetKey.self) { scrollOffset = $0 }
            }

            // ── 카드 탭 영역 ──
            Color.clear
                .frame(width: screenWidth, height: detailY + cardH + 28)
                .contentShape(Rectangle())
                .onTapGesture { onTapCard() }

            // ── 잔액 (고정 오버레이, 스크롤에 따라 fade + parallax) ──
            VStack(spacing: 6) {
                Text(balance.krwValue.krwFormatted)
                    .font(.system(size: 36, weight: .black, design: .rounded))
                    .foregroundStyle(.primary)
                Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
            .frame(maxWidth: .infinity)
            .frame(height: balanceH)
            .padding(.top, detailY + cardH + 28)
            .offset(y: balanceParallax)
            .opacity(balanceOpacity)
        }
        .frame(width: screenWidth)
        .transition(.opacity)
    }
}
