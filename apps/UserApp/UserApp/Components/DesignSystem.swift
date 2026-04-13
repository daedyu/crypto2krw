import SwiftUI

// MARK: - Color Extensions

extension Color {
    static let appPrimary    = Color(hex: "#5A4FCF")
    static let appBackground = Color(hex: "#FFFFFF")
    static let appSurface    = Color(hex: "#F2F2F7")
    static let appLabel      = Color(hex: "#1C1C1E")
    static let appSecondary  = Color(hex: "#8E8E93")
    static let appSeparator  = Color(hex: "#E5E5EA")
    static let appWarningBg  = Color(hex: "#FFF8E6")
    static let appWarning    = Color(hex: "#FF9500")
    static let appDanger     = Color(hex: "#FF3B30")

    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 6:
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8:
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (255, 0, 0, 0)
        }
        self.init(
            .sRGB,
            red:     Double(r) / 255,
            green:   Double(g) / 255,
            blue:    Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}

extension Currency {
    var accentColor: Color     { Color(hex: accentHex) }
    var backgroundColor: Color { Color(hex: backgroundHex) }

    var cardGradient: LinearGradient {
        switch self {
        case .USDT:
            return LinearGradient(
                colors: [Color(hex: "#1DB87E"), Color(hex: "#0A6644"), Color(hex: "#053D29")],
                startPoint: .topLeading, endPoint: .bottomTrailing
            )
        case .SOL:
            return LinearGradient(
                colors: [Color(hex: "#C060FF"), Color(hex: "#7B2FBE"), Color(hex: "#3D0A7A")],
                startPoint: .topLeading, endPoint: .bottomTrailing
            )
        case .ETH:
            return LinearGradient(
                colors: [Color(hex: "#6B8FF0"), Color(hex: "#2D5DD4"), Color(hex: "#0F2A8A")],
                startPoint: .topLeading, endPoint: .bottomTrailing
            )
        }
    }

    var shadowColor: Color {
        switch self {
        case .USDT: return Color(hex: "#1DB87E")
        case .SOL:  return Color(hex: "#9945FF")
        case .ETH:  return Color(hex: "#4A70E0")
        }
    }
}

// MARK: - Coin Badge

struct CoinBadge: View {
    let currency: Currency
    var size: CGFloat = 46

    var body: some View {
        ZStack {
            Circle()
                .fill(currency.accentColor)
                .frame(width: size, height: size)
            Text(String(currency.rawValue.prefix(1)))
                .font(.system(size: size * 0.4, weight: .black))
                .foregroundStyle(.white)
        }
    }
}

// MARK: - Section Card

struct SectionCard<Content: View>: View {
    @ViewBuilder let content: Content

    var body: some View {
        VStack(spacing: 0) {
            content
        }
        .background(Color.appBackground)
        .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
    }
}

// MARK: - Primary Button

struct PrimaryButton: View {
    let title: String
    let isEnabled: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(.system(size: 17, weight: .bold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 18)
                .background(isEnabled ? Color.appPrimary : Color.appPrimary.opacity(0.4))
                .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        }
        .disabled(!isEnabled)
    }
}

// MARK: - Input Field

struct AppTextField: View {
    let label: String
    let placeholder: String
    @Binding var text: String
    var isSecure: Bool = false
    var keyboardType: UIKeyboardType = .default

    @State private var isVisible = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(label)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Color(hex: "#3A3A3C"))

            ZStack(alignment: .trailing) {
                Group {
                    if isSecure && !isVisible {
                        SecureField(placeholder, text: $text)
                    } else {
                        TextField(placeholder, text: $text)
                            .keyboardType(keyboardType)
                            .autocorrectionDisabled()
                            .textInputAutocapitalization(.never)
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 16)
                .padding(.trailing, isSecure ? 44 : 0)
                .background(Color.appSurface)
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                .font(.system(size: 16))

                if isSecure {
                    Button {
                        isVisible.toggle()
                    } label: {
                        Image(systemName: isVisible ? "eye.slash" : "eye")
                            .foregroundStyle(Color.appSecondary)
                    }
                    .padding(.trailing, 14)
                }
            }
        }
    }
}

// MARK: - Number Formatters

extension Double {
    func clamped(to range: ClosedRange<Double>) -> Double {
        min(max(self, range.lowerBound), range.upperBound)
    }

    var krwFormatted: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .decimal
        formatter.maximumFractionDigits = 0
        return "₩" + (formatter.string(from: NSNumber(value: self)) ?? "0")
    }

    var coinFormatted: String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .decimal
        formatter.minimumFractionDigits = 0
        formatter.maximumFractionDigits = 6
        return formatter.string(from: NSNumber(value: self)) ?? "0"
    }
}

extension Date {
    var shortFormatted: String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.dateFormat = "M월 d일 HH:mm"
        return formatter.string(from: self)
    }

    var dateFormatted: String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ko_KR")
        formatter.dateFormat = "yyyy년 M월 d일"
        return formatter.string(from: self)
    }
}
