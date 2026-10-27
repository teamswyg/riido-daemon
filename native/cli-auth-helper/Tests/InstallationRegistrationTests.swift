import Foundation

func runInstallationRegistrationTests() {
    var events = [String]()
    var keys = [Data]()
    let registration = InstallationKeyRegistration(
        reserve: { events.append("reserve"); return InstallationKeyReservation(check: { events.append("check") }, release: { events.append("release") }) },
        load: { events.append("load"); return keys },
        create: { events.append("create"); let key = Data([1]); keys.append(key); return key },
        publicBytes: { $0 })
    check((try? registration.ensure()) == Data([1]), "first registration returns permanent key")
    check(events.first == "reserve" && events.last == "release", "registration holds reservation for find/create/recheck")
    check(events.filter { $0 == "load" }.count == 2, "created key rechecked under reservation")
    events = []
    check((try? registration.ensure()) == Data([1]) && !events.contains("create"), "later registration reuses key")
    keys.append(Data([2]))
    do { _ = try registration.ensure(); check(false, "ambiguous tag cannot select arbitrary key") } catch {}
    keys = []
    var released = false
    let failing = InstallationKeyRegistration<Data>(reserve: {
        InstallationKeyReservation(check: {}, release: { released = true })
    }, load: { [] }, create: { throw HelperFailure.unavailable }, publicBytes: { $0 })
    do { _ = try failing.ensure(); check(false, "creation failure refuses") } catch {}
    check(released, "creation failure releases reservation")

    final class Backend {
        let state = NSLock()
        let registration = NSLock()
        var values = [Data]()
        var results = [Data]()
        let firstCreate = DispatchSemaphore(value: 0)
        let allowCreate = DispatchSemaphore(value: 0)
        let secondLoad = DispatchSemaphore(value: 0)
        var loadCount = 0
        func load() -> [Data] {
            state.lock(); defer { state.unlock() }
            loadCount += 1
            if loadCount == 2 { secondLoad.signal() }
            return values
        }
        func create() -> Data {
            state.lock(); let first = values.isEmpty; state.unlock()
            if first { firstCreate.signal(); _ = allowCreate.wait(timeout: .now() + 2) }
            state.lock(); defer { state.unlock() }
            let key = Data([UInt8(values.count + 1)]); values.append(key); return key
        }
    }
    let backend = Backend()
    let concurrent = InstallationKeyRegistration(reserve: {
        backend.registration.lock()
        return InstallationKeyReservation(check: {}, release: { backend.registration.unlock() })
    }, load: backend.load, create: backend.create, publicBytes: { $0 })
    let group = DispatchGroup()
    func launch() {
        group.enter()
        DispatchQueue.global().async {
            let result = try? concurrent.ensure()
            backend.state.lock(); if let result { backend.results.append(result) }; backend.state.unlock()
            group.leave()
        }
    }
    launch()
    check(backend.firstCreate.wait(timeout: .now() + 1) == .success, "first creation reached backend")
    launch()
    check(backend.secondLoad.wait(timeout: .now() + 0.1) == .timedOut, "second registration cannot read through first creation")
    backend.allowCreate.signal(); backend.allowCreate.signal()
    check(group.wait(timeout: .now() + 3) == .success, "concurrent creation finishes")
    check(backend.values.count == 1 && backend.results == [Data([1]), Data([1])], "simultaneous first registration shares one key")
}
