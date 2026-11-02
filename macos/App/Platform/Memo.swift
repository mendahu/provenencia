/// The last value built, kept while its key is unchanged. A key holds every
/// input the build reads.
struct Memo<Key: Equatable, Value> {
    /// The stored key and value, or nil before the first store.
    private(set) var entry: (key: Key, value: Value)?

    /// The stored value when `key` matches the stored key.
    func value(for key: Key) -> Value? {
        guard let entry, entry.key == key else { return nil }
        return entry.value
    }

    mutating func store(_ value: Value, for key: Key) {
        entry = (key, value)
    }

    /// The stored value for `key`, else the value `build` returns, stored.
    mutating func value(for key: Key, build: () -> Value) -> Value {
        if let stored = value(for: key) { return stored }
        let built = build()
        store(built, for: key)
        return built
    }
}
