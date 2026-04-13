import CoreMotion
import SwiftUI

@Observable
final class MotionManager {
    var roll:  Double = 0
    var pitch: Double = 0

    private let manager = CMMotionManager()

    func start() {
        guard manager.isDeviceMotionAvailable else { return }
        manager.deviceMotionUpdateInterval = 1.0 / 60.0
        manager.startDeviceMotionUpdates(to: .main) { [weak self] data, _ in
            guard let self, let data else { return }
            withAnimation(.easeOut(duration: 0.12)) {
                self.roll  = data.attitude.roll
                self.pitch = data.attitude.pitch
            }
        }
    }

    func stop() {
        manager.stopDeviceMotionUpdates()
    }
}
