import SwiftUI
import CoreImage.CIFilterBuiltins

// MARK: - Home View

struct HomeView: View {
    @State private var selectedIndex: Int? = nil
    @State private var qrToken  = ""
    @State private var timeLeft = 30
    @State private var dragOffset: CGFloat = 0
    @State private var motion = MotionManager()

    private let balances    = MockData.balances
    private let cardGap: CGFloat  = 100   // 카드 간격 (넉넉하게)
    private let headerH: CGFloat  = 110   // 헤더 영역 높이
    private let selectedY: CGFloat = 24   // 선택 시 카드 상단 위치

    private var totalKrw: Double { balances.reduce(0) { $0 + $1.krwValue } }

    var body: some View {
        GeometryReader { geo in
            let cardW = geo.size.width - 48
            let cardH = cardW * 0.60

            ZStack(alignment: .topLeading) {

                // ── 헤더 ──
                VStack(alignment: .leading, spacing: 6) {
                    Text("내 지갑")
                        .font(.largeTitle.bold())
                    Text(totalKrw.krwFormatted)
                        .font(.title2)
                        .foregroundStyle(.secondary)
                }
                .padding(.horizontal, 24)
                .padding(.top, 8)
                .opacity(selectedIndex == nil ? 1 : 0)
                .animation(.easeInOut(duration: 0.2), value: selectedIndex)
                .zIndex(1)

                // ── 카드 스택 ──
                ForEach(Array(balances.enumerated()), id: \.element.id) { i, balance in
                    WalletCard(balance: balance, width: cardW, height: cardH, isSelected: selectedIndex == i,
                               roll: motion.roll, pitch: motion.pitch)
                        .offset(x: 24, y: yOffset(i: i, cardH: cardH, totalH: geo.size.height) + (selectedIndex == i ? dragOffset : 0))
                        .zIndex(selectedIndex == i ? 100 : Double(i))
                        .onTapGesture {
                            withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                                if selectedIndex == i {
                                    selectedIndex = nil
                                    dragOffset = 0
                                } else {
                                    selectedIndex = i
                                    dragOffset = 0
                                    refreshQR(i)
                                }
                            }
                        }
                }

                // ── QR 영역 (선택 시 표시) ──
                if let idx = selectedIndex {
                    VStack(spacing: 20) {
                        // 카드 아래에 위치하도록 상단 여백
                        Color.clear.frame(height: selectedY + cardH + 24)

                        // 잔액
                        VStack(spacing: 5) {
                            Text(balances[idx].krwValue.krwFormatted)
                                .font(.system(size: 30, weight: .black, design: .rounded))
                                .foregroundStyle(.primary)
                            Text("\(balances[idx].amount.coinFormatted) \(balances[idx].currency.rawValue)")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }

                        // QR 코드
                        QRCodeImage(content: qrToken)
                            .frame(width: 220, height: 220)
                            .padding(16)
                            .background(.background.secondary)
                            .clipShape(RoundedRectangle(cornerRadius: 20, style: .continuous))

                        // 타이머
                        VStack(spacing: 6) {
                            ProgressView(value: Double(timeLeft), total: 30.0)
                                .tint(balances[idx].currency.accentColor)
                                .frame(width: 200)
                            Text("\(timeLeft)초 후 자동 갱신")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }

                        Spacer()
                    }
                    .offset(y: dragOffset)
                    .frame(width: geo.size.width)
                    .transition(.opacity.combined(with: .move(edge: .bottom)))
                    .zIndex(50)
                }
            }
        }
        .navigationTitle("")
        .navigationBarTitleDisplayMode(.inline)
        .onAppear  { motion.start() }
        .onDisappear { motion.stop() }
        .simultaneousGesture(
            DragGesture(minimumDistance: 10)
                .onChanged { value in
                    guard selectedIndex != nil, value.translation.height > 0 else { return }
                    dragOffset = value.translation.height
                }
                .onEnded { value in
                    guard selectedIndex != nil else { return }
                    if value.translation.height > 100 {
                        withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                            selectedIndex = nil
                            dragOffset = 0
                        }
                    } else {
                        withAnimation(.spring(response: 0.4, dampingFraction: 0.85)) {
                            dragOffset = 0
                        }
                    }
                }
        )
        .onReceive(
            Timer.publish(every: 1, on: .main, in: .common).autoconnect()
        ) { _ in
            guard selectedIndex != nil else { return }
            if timeLeft <= 0 {
                if let idx = selectedIndex { refreshQR(idx) }
            } else {
                timeLeft -= 1
            }
        }
    }

    // MARK: - Helpers

    private func yOffset(i: Int, cardH: CGFloat, totalH: CGFloat) -> CGFloat {
        if let selected = selectedIndex {
            // 선택된 카드: 화면 상단으로
            if i == selected { return selectedY }
            // 나머지: 화면 아래로
            return totalH + 80
        }
        // 기본: 헤더 아래에 카드 간격으로 배치
        return headerH + CGFloat(i) * cardGap
    }

    private func refreshQR(_ idx: Int) {
        qrToken  = "crypto2krw://pay?coin=\(balances[idx].currency.rawValue)&ts=\(Int(Date().timeIntervalSince1970))"
        timeLeft = 30
    }
}

// MARK: - Wallet Card

struct WalletCard: View {
    let balance:    CoinBalance
    let width:      CGFloat
    let height:     CGFloat
    var isSelected: Bool   = false
    var roll:       Double = 0
    var pitch:      Double = 0

    // 기울기 → 하이라이트 UnitPoint (중앙 기준, ±0.5 범위로 클램프)
    private var highlightCenter: UnitPoint {
        let sensitivity = 0.45
        let x = (0.25 - roll  * sensitivity).clamped(to: 0...1)
        let y = (0.25 + pitch * sensitivity).clamped(to: 0...1)
        return UnitPoint(x: x, y: y)
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            RoundedRectangle(cornerRadius: 22, style: .continuous)
                .fill(balance.currency.cardGradient)

            // 기울기 반응 하이라이트
            RoundedRectangle(cornerRadius: 22, style: .continuous)
                .fill(
                    RadialGradient(
                        colors: [.white.opacity(0.18), .clear],
                        center: highlightCenter,
                        startRadius: 0,
                        endRadius: width * 0.75
                    )
                )

            VStack(alignment: .leading, spacing: 0) {
                // 상단: 코인명 + 심볼
                HStack(alignment: .top) {
                    VStack(alignment: .leading, spacing: 3) {
                        Text(balance.currency.rawValue)
                            .font(.system(size: 20, weight: .black, design: .rounded))
                            .foregroundStyle(.white)
                        Text(balance.currency.fullName)
                            .font(.system(size: 12, weight: .medium))
                            .foregroundStyle(.white.opacity(0.6))
                    }
                    Spacer()
                    ZStack {
                        Circle()
                            .fill(.white.opacity(0.12))
                            .frame(width: 40, height: 40)
                        Text(balance.currency.symbol)
                            .font(.system(size: 18, weight: .bold))
                            .foregroundStyle(.white)
                    }
                }

                Spacer()

                // 하단: 원화 가치 + 수량
                VStack(alignment: .leading, spacing: 4) {
                    Text(balance.krwValue.krwFormatted)
                        .font(.system(size: 28, weight: .black, design: .rounded))
                        .foregroundStyle(.white)
                        .minimumScaleFactor(0.7)
                        .lineLimit(1)
                    Text("\(balance.amount.coinFormatted) \(balance.currency.rawValue)")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(.white.opacity(0.6))
                }
            }
            .padding(22)
        }
        .frame(width: width, height: height)
        .rotation3DEffect(
            .degrees(isSelected ? 0 : -14),
            axis: (x: 1, y: 0, z: 0),
            anchor: .center,
            perspective: 0.4
        )
        .shadow(
            color: balance.currency.shadowColor.opacity(0.4),
            radius: 22, x: 0, y: 12
        )
    }
}

// MARK: - QR Code (CoreImage 네이티브)

struct QRCodeImage: View {
    let content: String

    var body: some View {
        if let uiImage = makeQR() {
            Image(uiImage: uiImage)
                .interpolation(.none)
                .resizable()
                .scaledToFit()
        } else {
            ProgressView()
        }
    }

    private func makeQR() -> UIImage? {
        guard !content.isEmpty else { return nil }
        let ctx    = CIContext()
        let filter = CIFilter.qrCodeGenerator()
        filter.message         = Data(content.utf8)
        filter.correctionLevel = "M"
        guard let out = filter.outputImage else { return nil }
        let scaled = out.transformed(by: CGAffineTransform(scaleX: 12, y: 12))
        guard let cg = ctx.createCGImage(scaled, from: scaled.extent) else { return nil }
        return UIImage(cgImage: cg)
    }
}

#Preview {
    NavigationStack {
        HomeView()
            .environment(AuthManager())
    }
}
