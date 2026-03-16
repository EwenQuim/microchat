# React Native Refactor Evaluation for Microchat

## Summary

This document evaluates refactoring microchat to React Native + React Native Web for mobile app support.

## Current Architecture

| Layer | Technology |
|-------|------------|
| Framework | React 19.2, TypeScript 5.7 |
| Build | Vite 7.1, PWA with service worker |
| Routing | TanStack Router (file-based) |
| State | TanStack Query v5 |
| Styling | Tailwind CSS v4, Radix UI (shadcn pattern) |
| Auth | Nostr-style cryptographic signing (secp256k1) |

## Recommendation

**Approach: Expo + NativeWind in a Turborepo monorepo**

- **Expo** (not bare RN): Crypto libraries are pure JS, no native modules needed
- **NativeWind v4**: Preserves existing Tailwind class patterns
- **Expo Router**: Mirrors current file-based routing
- **Keep web app as-is**: RN Web adds complexity without clear benefit for this app

## What Can Be Reused (~60-70%)

| Code | Reusability |
|------|-------------|
| `@noble/secp256k1`, `@noble/hashes`, `@scure/base` | 100% - Pure JS crypto |
| `@tanstack/react-query` hooks | 100% - Works unchanged |
| API layer (Orval-generated) | 95% - Need base URL config |
| Business logic in hooks | 90% - Need storage abstraction |
| date-fns, clsx, tailwind-merge | 100% |

## What Must Be Replaced

| Web Dependency | React Native Alternative |
|----------------|-------------------------|
| `@radix-ui/*` | React Native core + custom components |
| `lucide-react` | `@expo/vector-icons` |
| `cmdk` | Custom search modal or omit on mobile |
| `qrcode` | `react-native-qrcode-svg` |
| `localStorage` | `@react-native-async-storage/async-storage` |
| Private key storage | `expo-secure-store` (security improvement) |
| `navigator.share()` | `expo-sharing` / RN Share |
| `navigator.clipboard` | `expo-clipboard` |
| DOM keyboard events | Platform-specific handling |

## Proposed Project Structure

```
microchat/
├── apps/
│   ├── mobile/                 # New Expo app
│   │   ├── app/                # Expo Router screens
│   │   │   ├── _layout.tsx
│   │   │   ├── index.tsx
│   │   │   ├── chat/[roomName].tsx
│   │   │   ├── settings.tsx
│   │   │   └── user/[pubkey].tsx
│   │   └── components/
│   └── web/                    # Existing Vite app (renamed from app/)
├── packages/
│   └── shared/
│       ├── crypto/             # Extract from app/src/lib/crypto.ts
│       ├── api/                # Orval types + fetch wrapper
│       ├── hooks/              # Shared hooks (useRooms, useMessages, etc.)
│       └── storage/            # Platform-agnostic storage interface
└── turbo.json
```

## Implementation Phases

### Phase 1: Monorepo Setup (1 week)
- Initialize Turborepo at project root
- Create `packages/shared/` with extracted crypto, types, storage interface
- Set up Expo app with Expo Router and NativeWind
- Verify crypto operations work on iOS/Android simulators

### Phase 2: Core Infrastructure (1-2 weeks)
- Implement storage abstraction layer:
  - `AsyncStorage` for general data
  - `expo-secure-store` for private keys
- Migrate hooks to shared package with platform-agnostic storage
- Configure API client with mobile-appropriate base URL

### Phase 3: Screen Migration (2-3 weeks)
- **Home/Rooms**: FlatList-based room list with pull-to-refresh
- **Chat**: Inverted FlatList for messages, TextInput for compose
- **Settings**: Tab-based view with QR code display
- **User Profile**: Basic profile view
- Replace Radix UI modals with React Native Modal or bottom sheets

### Phase 4: Native Features (1-2 weeks)
- Share functionality via `expo-sharing`
- Clipboard via `expo-clipboard`
- Deep linking for profile import (`microchat://settings?import=nsec1...`)
- Optional: Push notifications, QR code scanning

### Phase 5: Testing & Polish (1-2 weeks)
- Platform-specific testing (iOS safe areas, Android back button)
- Performance optimization (FlatList tuning)
- App Store/Play Store preparation

## Effort Estimate

**Total: 6-10 weeks** for a senior developer

| Phase | Effort | Complexity |
|-------|--------|------------|
| Setup | 1 week | Low |
| Infrastructure | 1-2 weeks | Medium |
| Screen Migration | 2-3 weeks | High |
| Native Features | 1-2 weeks | Medium |
| Testing/Polish | 1-2 weeks | Medium |

## Key Risks

| Risk | Mitigation |
|------|------------|
| Storage migration | Comprehensive tests, data migration path |
| Crypto compatibility | Test on physical devices early (pure JS should work) |
| UI parity | Accept some mobile-specific adaptations (no cmd+K) |
| Maintenance burden | Clear shared/platform-specific separation |

## React Native Web Decision

**Recommendation: Do not use React Native Web initially**

Reasons:
- Existing Vite web app is optimized and has PWA support
- RN Web adds bundle size and complexity
- Different UX expectations for web vs mobile
- Can evaluate later if dual maintenance becomes costly

## Critical Files

Files requiring migration or abstraction:
- `app/src/lib/storage.ts` - Platform-agnostic interface
- `app/src/lib/crypto.ts` - Extract to shared package
- `app/src/hooks/useUsername.ts` - Use storage abstraction
- `app/src/hooks/useRoomPassword.ts` - Use storage abstraction
- `app/src/components/chat/ChatLayout.tsx` - Redesign for mobile nav

## Verification Plan

1. **Unit tests**: Shared crypto and hook logic
2. **Simulator testing**: iOS and Android for each screen
3. **Physical device testing**: Verify crypto signing works correctly
4. **Deep link testing**: Profile import via URL scheme
5. **Comparison testing**: Feature parity with web app

## Alternatives Considered

| Option | Verdict |
|--------|---------|
| React Native Web for all | Too complex, loses PWA benefits |
| Capacitor/Ionic | WebView performance, not truly native |
| Flutter | Complete rewrite, loses React ecosystem |
| Keep web-only PWA | Viable but misses native app store presence |
