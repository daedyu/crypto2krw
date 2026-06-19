import SwiftUI
import CoreImage.CIFilterBuiltins

// MARK: - Home ViewModel

@Observable
final class HomeViewModel {
    var balances: [CoinBalance] = []
    var transactions: [Transaction] = []
    var isLoading = false
    var errorMessage: String? = nil

    @MainActor
    func load() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let b  = UserService.shared.getBalances()
            async let w  = UserService.shared.getWallets()
            async let r  = UserService.shared.getRates()
            async let tx = UserService.shared.getTransactions(limit: 100)
            let (bals, walls, rates, txItems) = try await (b, w, r, tx)
            transactions = txItems.compactMap { $0.toTransaction() }
            balances = Self.buildCoinBalances(balances: bals, wallets: walls, rates: rates)
        } catch {
            errorMessage = (error as? APIClientError)?.errorDescription ?? error.localizedDescription
        }
    }

    private static func buildCoinBalances(
        balances: [BalanceItem],
        wallets: [WalletItem],
        rates: RateMap
    ) -> [CoinBalance] {
        let rateMap: [Currency: Double] = [
            .SOL:  Double(rates.sol)  ?? 0,
            .ETH:  Double(rates.eth)  ?? 0,
            .USDT: Double(rates.usdt) ?? 0,
        ]
        var amounts: [Currency: Double] = [:]
        for item in balances {
            guard let c = Currency(apiValue: item.currency) else { continue }
            amounts[c, default: 0] += Double(item.availableBalance) ?? 0
        }
        var primaryWallet: [Currency: WalletItem] = [:]
        for w in wallets.sorted(by: { $0.paymentPriority < $1.paymentPriority }) {
            guard let c = Currency(apiValue: w.currency), primaryWallet[c] == nil else { continue }
            primaryWallet[c] = w
        }
        return [Currency.USDT, .SOL, .ETH].compactMap { c in
            guard let wallet = primaryWallet[c] else { return nil }
            let amount = amounts[c] ?? 0
            return CoinBalance(currency: c, amount: amount,
                               krwValue: amount * (rateMap[c] ?? 0),
                               address: wallet.address)
        }
    }
}

// MARK: - Home View

struct HomeView: View {
    @State private var selectedIndex: Int? = nil
    @State private var showDetail = false
    @State private var qrToken   = ""
    @State private var timeLeft  = 30
    @State private var dragOffset: CGFloat = 0
    @State private var motion = MotionManager()
    @State private var vm = HomeViewModel()

    // 결제 감지
    @State private var prevAmounts: [Currency: Double] = [:]
    @State private var paymentDone   = false   // 결제 완료 오버레이

    private let cardGap: CGFloat   = 110
    private let headerH: CGFloat   = 110
    private let selectedY: CGFloat = 24
    private let detailY:   CGFloat = 60
    @State private var detailContentVisible = false

    private var balances: [CoinBalance] { vm.balances }
    private var totalKrw: Double { vm.balances.reduce(0) { $0 + $1.krwValue } }

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
                    WalletCard(
                        balance: balance, width: cardW, height: cardH,
                        isSelected: selectedIndex == i,
                        roll: motion.roll, pitch: motion.pitch
                    )
                    .offset(
                        x: 24,
                        y: yOffset(i: i, cardH: cardH, totalH: geo.size.height)
                            + (selectedIndex == i && !showDetail ? dragOffset : 0)
                    )
                    .animation(
                        .spring(response: 0.55, dampingFraction: 0.82),
                        value: showDetail
                    )
                    .zIndex(selectedIndex == i ? 100 : Double(i))
                    .opacity(showDetail && selectedIndex != i ? 0 : 1)
                    .onTapGesture {
                        if selectedIndex == i {
                            if showDetail {
                                withAnimation(.spring(response: 0.55, dampingFraction: 0.82)) {
                                    showDetail = false
                                    detailContentVisible = false
                                }
                            } else {
                                withAnimation(.spring(response: 0.55, dampingFraction: 0.82)) {
                                    showDetail = true
                                    detailContentVisible = false
                                    dragOffset = 0
                                }
                                DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                                    withAnimation(.easeOut(duration: 0.22)) {
                                        detailContentVisible = true
                                    }
                                }
                            }
                        } else {
                            withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                                selectedIndex = i
                                showDetail = false
                                dragOffset = 0
                                paymentDone = false
                                initPollSnapshot()
                                refreshQR(i)
                            }
                        }
                    }
                }

                // ── QR 영역 ──
                if let idx = selectedIndex, !showDetail {
                    VStack(spacing: 20) {
                        Color.clear.frame(height: selectedY + cardH + 24)

                        VStack(spacing: 5) {
                            Text(balances[idx].krwValue.krwFormatted)
                                .font(.system(size: 30, weight: .black, design: .rounded))
                            Text("\(balances[idx].amount.coinFormatted) \(balances[idx].currency.rawValue)")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }

                        // QR 또는 결제완료 오버레이
                        ZStack {
                            QRCodeImage(content: qrToken)
                                .frame(width: 220, height: 220)
                                .opacity(paymentDone ? 0 : 1)

                            if paymentDone {
                                PaymentSuccessOverlay()
                                    .frame(width: 220, height: 220)
                                    .transition(.scale.combined(with: .opacity))
                            }
                        }
                        .padding(16)
                        .background(.background.secondary)
                        .clipShape(RoundedRectangle(cornerRadius: 20, style: .continuous))
                        .animation(.spring(response: 0.4, dampingFraction: 0.7), value: paymentDone)

                        if !paymentDone {
                            VStack(spacing: 6) {
                                ProgressView(value: Double(timeLeft), total: 30.0)
                                    .tint(balances[idx].currency.accentColor)
                                    .frame(width: 200)
                                Text("\(timeLeft)초 후 자동 갱신")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }

                        Spacer()
                    }
                    .offset(y: dragOffset)
                    .frame(width: geo.size.width)
                    .transition(.opacity.combined(with: .move(edge: .bottom)))
                    .zIndex(50)
                }

                // ── 디테일 뷰 ──
                if showDetail, detailContentVisible, let idx = selectedIndex {
                    CardDetailView(
                        balance: balances[idx],
                        cardH: cardH,
                        detailY: detailY,
                        screenWidth: geo.size.width,
                        transactions: vm.transactions.filter { $0.usedCurrency == balances[idx].currency }
                    ) {
                        withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                            showDetail = false
                            detailContentVisible = false
                        }
                    }
                    .frame(width: geo.size.width, height: geo.size.height)
                    .zIndex(150)
                }
            }
        }
        .navigationTitle(showDetail ? (selectedIndex.map { balances[$0].currency.fullName } ?? "") : "")
        .navigationBarTitleDisplayMode(.inline)
        // 뷰가 나타날 때마다 항상 최신 데이터 로드
        .task { await vm.load() }
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                if showDetail {
                    Button {
                        withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                            showDetail = false
                            detailContentVisible = false
                            selectedIndex = nil
                            dragOffset = 0
                        }
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundStyle(.secondary)
                            .font(.title3)
                    }
                    .transition(.opacity)
                }
            }
        }
        .animation(.easeInOut(duration: 0.2), value: showDetail)
        .onAppear   { motion.start() }
        .onDisappear { motion.stop() }
        .simultaneousGesture(
            DragGesture(minimumDistance: 10)
                .onChanged { value in
                    guard !showDetail, selectedIndex != nil,
                          value.translation.height > 0 else { return }
                    dragOffset = value.translation.height
                }
                .onEnded { value in
                    if showDetail {
                        if value.translation.height > 80 {
                            withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                                showDetail = false
                                detailContentVisible = false
                                selectedIndex = nil
                                dragOffset = 0
                            }
                        }
                        return
                    }
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
        .onReceive(Timer.publish(every: 1, on: .main, in: .common).autoconnect()) { _ in
            guard selectedIndex != nil, !showDetail else { return }

            // QR 갱신 카운트다운
            if timeLeft <= 0 {
                if let idx = selectedIndex { refreshQR(idx) }
            } else {
                timeLeft -= 1
            }

            // 1초마다 잔액 폴링 — 결제 완료 감지
            Task { await detectPayment() }
        }
    }

    // MARK: - QR 갱신

    private func refreshQR(_ idx: Int) {
        let coin = balances[idx].currency.rawValue
        let at   = TokenStore.accessToken ?? ""
        qrToken  = at.isEmpty
            ? "crypto2krw://pay?coin=\(coin)&ts=\(Int(Date().timeIntervalSince1970))"
            : "crypto2krw://pay?coin=\(coin)&at=\(at)"
        timeLeft = 30
    }

    // MARK: - 결제 감지

    private func initPollSnapshot() {
        for b in balances { prevAmounts[b.currency] = b.amount }
    }

    @MainActor
    private func detectPayment() async {
        guard let idx = selectedIndex, !showDetail, !paymentDone else { return }
        let currency = balances[idx].currency
        guard let prev = prevAmounts[currency] else { return }

        guard let freshItems = try? await UserService.shared.getBalances() else { return }

        var newAmount: Double? = nil
        for item in freshItems {
            guard let c = Currency(apiValue: item.currency), c == currency else { continue }
            newAmount = (newAmount ?? 0) + (Double(item.availableBalance) ?? 0)
        }

        guard let fresh = newAmount else { return }

        // 잔액 감소 = 결제 완료
        if fresh < prev - 0.000001 {
            UINotificationFeedbackGenerator().notificationOccurred(.success)

            withAnimation(.spring(response: 0.4, dampingFraction: 0.7)) {
                paymentDone = true
            }

            // 전체 데이터 갱신
            await vm.load()

            // 3초 후 카드 닫기
            try? await Task.sleep(nanoseconds: 3_000_000_000)
            withAnimation(.spring(response: 0.5, dampingFraction: 0.78)) {
                selectedIndex = nil
                paymentDone   = false
                dragOffset    = 0
            }
        } else {
            prevAmounts[currency] = fresh
        }
    }

    // MARK: - Layout

    private func yOffset(i: Int, cardH: CGFloat, totalH: CGFloat) -> CGFloat {
        if let selected = selectedIndex {
            if i == selected { return showDetail ? detailY : selectedY }
            return totalH + 70
        }
        return headerH + CGFloat(i) * cardGap
    }
}

// MARK: - 결제 완료 오버레이

struct PaymentSuccessOverlay: View {
    @State private var scale = 0.3
    @State private var opacity = 0.0

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 20, style: .continuous)
                .fill(Color.green.opacity(0.15))

            Image(systemName: "checkmark")
                .font(.system(size: 72, weight: .semibold))
                .foregroundStyle(.green)
                .scaleEffect(scale)
                .opacity(opacity)
        }
        .onAppear {
            withAnimation(.spring(response: 0.4, dampingFraction: 0.55)) {
                scale   = 1.0
                opacity = 1.0
            }
        }
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

    private var highlightCenter: UnitPoint {
        let s = 0.45
        let x = (0.25 - roll  * s).clamped(to: 0...1)
        let y = (0.25 + pitch * s).clamped(to: 0...1)
        return UnitPoint(x: x, y: y)
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            RoundedRectangle(cornerRadius: 22, style: .continuous)
                .fill(balance.currency.cardGradient)

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

// MARK: - QR Code

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
