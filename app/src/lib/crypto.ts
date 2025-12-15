import { sha256 } from "@noble/hashes/sha2.js";
import { bytesToHex, hexToBytes } from "@noble/hashes/utils.js";
import * as secp256k1 from "@noble/secp256k1";

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
