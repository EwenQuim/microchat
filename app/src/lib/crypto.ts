import { sha256 } from "@noble/hashes/sha2.js";
import { bytesToHex, hexToBytes } from "@noble/hashes/utils.js";
import * as secp256k1 from "@noble/secp256k1";
import { bech32 } from "@scure/base";

/**
 * Nostr-style cryptographic utilities for message signing and verification
 * Following the Nostr protocol (NIPs) for key generation and signing
 */

export interface KeyPair {
	privateKey: string; // hex-encoded private key
	publicKey: string; // hex-encoded public key (Nostr npub format without prefix)
}

/**
 * Generate a new secp256k1 key pair for user identity
 * Returns hex-encoded private and public keys
 */
export async function generateKeyPair(): Promise<KeyPair> {
	const privateKey = secp256k1.utils.randomSecretKey();
	const publicKey = secp256k1.getPublicKey(privateKey);

	return {
		privateKey: bytesToHex(privateKey),
		publicKey: bytesToHex(publicKey),
	};
}

/**
 * Create a canonical event hash for signing
 * This follows Nostr's event serialization format
 */
function createEventHash(params: {
	publicKey: string;
	timestamp: number;
	content: string;
	room: string;
}): string {
	// Nostr event format: [0, pubkey, created_at, kind, tags, content]
	// We simplify by using: [0, pubkey, created_at, content, room]
	const serialized = JSON.stringify([
		0, // reserved for future use
		params.publicKey,
		params.timestamp,
		params.content,
		params.room,
	]);

	const hash = sha256(new TextEncoder().encode(serialized));
	return bytesToHex(hash);
}

/**
 * Sign a message using the private key
 * Returns the signature as a hex string
 */
export async function signMessage(params: {
	privateKey: string;
	publicKey: string;
	content: string;
	room: string;
	timestamp?: number;
}): Promise<{ signature: string; timestamp: number; eventHash: string }> {
	const timestamp = params.timestamp || Math.floor(Date.now() / 1000);

	// Create the event hash
	const eventHash = createEventHash({
		publicKey: params.publicKey,
		timestamp,
		content: params.content,
		room: params.room,
	});

	// Sign the hash
	const privateKeyBytes = hexToBytes(params.privateKey);
	const eventHashBytes = hexToBytes(eventHash);

	// signAsync returns compact signature bytes (64 bytes: 32-byte R + 32-byte S) by default
	// IMPORTANT: Use prehash: false because eventHashBytes is already a SHA-256 hash
	const signatureBytes = await secp256k1.signAsync(
		eventHashBytes,
		privateKeyBytes,
		{
			prehash: false,
		},
	);
	const signatureHex = bytesToHex(signatureBytes);

	// Self-verify the signature immediately
	const selfVerify = secp256k1.verify(
		signatureBytes,
		eventHashBytes,
		hexToBytes(params.publicKey),
		{ prehash: false },
	);

	// Debug logging
	console.log("DEBUG Frontend Signing:");
	console.log("  Event hash:", eventHash);
	console.log("  Signature length:", signatureBytes.length, "bytes");
	console.log("  Signature hex length:", signatureHex.length, "chars");
	console.log("  Self-verification:", selfVerify ? "✓ PASS" : "✗ FAIL");
	console.log("  Content:", params.content);
	console.log("  Room:", params.room);
	console.log("  Timestamp:", timestamp);
	console.log("  Pubkey:", params.publicKey);

	if (!selfVerify) {
		console.error("WARNING: Signature self-verification failed!");
	}

	return {
		signature: signatureHex,
		timestamp,
		eventHash,
	};
}

/**
 * Verify a message signature
 * Returns true if the signature is valid
 * @public
 */
export async function verifySignature(params: {
	publicKey: string;
	signature: string;
	content: string;
	room: string;
	timestamp: number;
}): Promise<boolean> {
	try {
		const eventHash = createEventHash({
			publicKey: params.publicKey,
			timestamp: params.timestamp,
			content: params.content,
			room: params.room,
		});

		const publicKeyBytes = hexToBytes(params.publicKey);
		const signatureBytes = hexToBytes(params.signature);
		const eventHashBytes = hexToBytes(eventHash);

		// Verify using compact signature bytes (64 bytes)
		// Use prehash: false because we're passing the hash directly (already hashed)
		return secp256k1.verify(signatureBytes, eventHashBytes, publicKeyBytes, {
			prehash: false,
		});
	} catch (error) {
		console.error("Signature verification failed:", error);
		return false;
	}
}

/**
 * Get a shortened version of a public key for display
 * e.g., "npub1abc...xyz"
 * @public
 */
export function formatPublicKey(publicKey: string): string {
	if (publicKey.length < 16) return publicKey;
	return `npub1${publicKey.slice(0, 8)}...${publicKey.slice(-8)}`;
}

/**
 * Convert a hex public key to npub (Nostr bech32 format)
 * @param hexPubKey - The hex-encoded public key (66 chars with 04 prefix or 64 chars without)
 * @returns The bech32-encoded npub string
 */
export function hexToNpub(hexPubKey: string): string {
	// Remove the 04 prefix if it exists (uncompressed public key indicator)
	let cleanHex = hexPubKey;
	if (hexPubKey.startsWith("04") && hexPubKey.length === 130) {
		// For uncompressed keys, we only use the x-coordinate (first 32 bytes after 04)
		cleanHex = hexPubKey.slice(2, 66);
	} else if (hexPubKey.length === 66) {
		// Remove any prefix (02, 03, 04)
		cleanHex = hexPubKey.slice(2);
	}

	const bytes = hexToBytes(cleanHex);
	const words = bech32.toWords(bytes);
	return bech32.encode("npub", words, 5000);
}

/**
 * Convert a hex private key to nsec (Nostr bech32 format)
 * @param hexPrivKey - The hex-encoded private key
 * @returns The bech32-encoded nsec string
 */
export function hexToNsec(hexPrivKey: string): string {
	const bytes = hexToBytes(hexPrivKey);
	const words = bech32.toWords(bytes);
	return bech32.encode("nsec", words, 5000);
}

/**
 * Convert an npub (Nostr bech32 format) to hex public key
 * @param npub - The bech32-encoded npub string
 * @returns The hex-encoded public key
 */
export function npubToHex(npub: string): string {
	const decoded = bech32.decode(npub as `${string}1${string}`, 5000);
	if (decoded.prefix !== "npub") {
		throw new Error("Invalid npub format");
	}
	const bytes = bech32.fromWords(decoded.words);
	return bytesToHex(new Uint8Array(bytes));
}

/**
 * Convert an nsec (Nostr bech32 format) to hex private key
 * @param nsec - The bech32-encoded nsec string
 * @returns The hex-encoded private key
 */
export function nsecToHex(nsec: string): string {
	const decoded = bech32.decode(nsec as `${string}1${string}`, 5000);
	if (decoded.prefix !== "nsec") {
		throw new Error("Invalid nsec format");
	}
	const bytes = bech32.fromWords(decoded.words);
	return bytesToHex(new Uint8Array(bytes));
}

/**
 * Derive public key from private key
 * @param privateKeyHex - The hex-encoded private key
 * @returns The hex-encoded public key
 */
export function derivePublicKey(privateKeyHex: string): string {
	const privateKeyBytes = hexToBytes(privateKeyHex);
	const publicKeyBytes = secp256k1.getPublicKey(privateKeyBytes);
	return bytesToHex(publicKeyBytes);
}
