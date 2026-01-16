# REVIEW.md - Version-Aware Identity System

## Problems That Existed

### 1. Static Identity Modeling
**Before**: "Chrome on Android" was treated as one monolithic identity.
- All Android sessions looked identical
- No accounting for Android version diversity (10, 11, 12, 13, 14)
- No accounting for Chrome version diversity (114-121)

**Why This Matters**: Anti-bot systems can detect uniform fingerprints at scale. If 10,000 requests all claim "Android 10, Chrome 120", despite these being minority versions, it's suspicious.

### 2. User-Agent Unrealism
**Before**: User-Agent strings were static templates.
- No version variation
- No correlation between claimed browser version and actual behavior
- Missed subtle differences (e.g., Android 10 vs 13 UA format)

**Why This Matters**: Real user populations have version diversity. A scraper that doesn't model this looks artificial under statistical analysis.

### 3. TLS Fingerprint Mismatch Risk
**Before**: TLS fingerprints were loosely coupled to browser identity.
- Chrome 120 fingerprint might be used for all "Chrome" requests
- No consideration of version evolution in TLS stacks

**Why This Matters**: Claiming "Chrome 115" in User-Agent while using a Chrome 120 TLS fingerprint is a detectable inconsistency.

### 4. Missing Header Correlation
**Before**: `sec-ch-ua` and other headers were static.
- `sec-ch-ua` might claim Chrome 120 while UA says Chrome 118
- No version alignment between header values

**Why This Matters**: Modern browsers send version info in multiple headers. Inconsistencies are red flags.

## What Was Changed

### 1. Version Type System (`types/versions.go`)
- Added `BrowserVersion` struct (major, minor, patch)
- Added `OSVersion` struct (flexible for different OS schemes)
- Defined realistic version ranges (Chrome 114-121, Android 10-14, etc.)
- Documented compatibility rules

### 2. User-Agent Generation (`identity/useragent.go`)
- Completely rewrote UA generation with version-aware templates
- **Android**: UA now reflects specific Android version (10, 11, 12, 13, 14)
- **iOS**: UA reflects iOS version with proper underscore format
- **Desktop**: Windows, macOS, Linux variations with OS versions
- Detailed comments explaining format differences

### 3. Version Selection Logic (`identity/versions.go`)
- Implemented smart version picking within realistic ranges
- Chrome: Random selection from 114-121
- Safari: Random selection from 15-16
- Android OS: Random selection from 10-14
- Compatibility validation (e.g., Chrome 120 requires Android 10+)

### 4. TLS Fingerprint Mapping (`identity/fingerprints.go`)
- Created version-aware TLS fingerprint selection
- Chrome 120+ → Chrome 120 fingerprint
- Chrome 106-119 → Chrome 106 fingerprint (closest match)
- Safari 15-16 → Safari 16 fingerprint
- Documented approximations where exact mapping unavailable

### 5. Identity Struct Enhancement (`identity/identity.go`)
- Added `BrowserVersion` and `OSVersion` fields
- Identity now represents a complete, versioned browser instance
- Immutable once assigned to a session

### 6. Generator Refactor (`identity/generator.go`)
- Integrated version selection into identity generation flow
- Ensured derived values (UA, sec-ch-ua, TLS) align with versions
- Added constraint validation (no impossible combinations)

## Why These Changes Matter

### Statistical Realism
**Before**: All mobile Chrome sessions were identical → easy to fingerprint the scraper.  
**After**: Mobile Chrome sessions vary across Android 10-14, Chrome 114-121 → looks like real user population.

### Consistency Within Sessions
**Before**: No guarantee that headers matched each other.  
**After**: User-Agent version == sec-ch-ua version == TLS fingerprint version.

### Controlled Entropy
**Before**: Either too static or too random.  
**After**: Randomness constrained by real-world compatibility rules.

## Remaining Limitations

### 1. TLS Fingerprint Granularity
**Issue**: uTLS doesn't provide fingerprints for every browser version.  
**Mitigation**: We map version ranges to closest available fingerprints.  
**Example**: Chrome 114-119 all use Chrome 106 fingerprint.

**Why Acceptable**:
- TLS stacks don't change every version
- Minor version TLS differences are very subtle
- Consistency within a session is more important than exact matching

### 2. Header Ordering
**Issue**: Go's `http.Header` is a map, which has undefined iteration order.  
**Mitigation**: Standard library writes headers in a stable order for common keys.  
**Future**: Could implement custom RoundTripper for byte-perfect ordering.

### 3. No JavaScript Execution
**Limitation**: Some sites require JS-calculated tokens, cookies, or challenges.  
**Not Planned**: This would require a full browser engine, violating the HTTP-only constraint.

### 4. Browser Version Staleness
**Issue**: Our version ranges (Chrome 114-121) will become outdated.  
**Mitigation**: These are configurable constants in `types/versions.go`.  
**Future**: Could auto-update from external data source.

## Design Decisions Explained

### Decision 1: Version Ranges vs Exact Versions
**Choice**: Use ranges (Chrome 114-121) rather than exact versions.  
**Rationale**:
- Real users don't all update immediately
- Version diversity is realistic
- Easier to maintain (update range periodically vs tracking every release)

### Decision 2: Approximation in TLS Mapping
**Choice**: Map Chrome 114-119 → Chrome 106 fingerprint.  
**Rationale**:
- uTLS doesn't have fingerprints for every version
- TLS stacks are stable across minor versions
- Session consistency more important than exact matching
- Documented clearly for transparency

### Decision 3: Static Android "K" Device Model
**Choice**: Use "K" for all Android 10+ devices.  
**Rationale**:
- Google introduced "K" as device model obfuscation in Android 10
- Real devices now show "K" instead of specific models
- Simplifies implementation while remaining realistic

### Decision 4: Immutable Session Identities
**Choice**: Sessions lock their identity (versions included).  
**Rationale**:
- Real users don't change OS/browser version mid-session
- Version drift would be highly suspicious
- Explicit rotation is allowed but not automatic

## Future Enhancements

1. **Identity Pool**: Pre-generate a pool of N identities, distribute round-robin.
2. **Firefox Support**: Add Firefox version ranges and fingerprints.
3. **Auto-Version Updates**: Fetch latest browser versions from CDN.
4. **Fine-Grained Header Ordering**: Custom RoundTripper for exact header order.
5. **Locale Diversity**: Model realistic Accept-Language distributions by region.

## Conclusion

The version-aware identity system transforms goscraping from a feature-based scraper into a **statistical fingerprinting system**.

The goal is not perfect simulation of any single browser, but realistic distribution across a population of browsers—just like real web traffic.
