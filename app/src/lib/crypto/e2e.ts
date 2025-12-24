/**
 * End-to-End Encryption utilities using Web Crypto API
 *
 * Uses AES-GCM for authenticated encryption and PBKDF2 for key derivation.
 */

const PBKDF2_ITERATIONS = 100000;
const KEY_LENGTH = 256; // bits
const NONCE_LENGTH = 12; // bytes (96 bits for AES-GCM)
const SALT_LENGTH = 32; // bytes

/**
 * Converts a hex string to Uint8Array
 */
function hexToBytes(hex: string): Uint8Array {
	const bytes = new Uint8Array(hex.length / 2);
	for (let i = 0; i < hex.length; i += 2) {
		bytes[i / 2] = parseInt(hex.slice(i, i + 2), 16);
	}
	return bytes;
}

/**
 * Converts Uint8Array to hex string
 */
function bytesToHex(bytes: Uint8Array): string {
	return Array.from(bytes)
		.map((b) => b.toString(16).padStart(2, "0"))
		.join("");
}

/**
 * Derives an AES-GCM encryption key from a password using PBKDF2
 *
 * @param password - The password to derive the key from
 * @param salt - Hex-encoded salt (must be consistent per room)
 * @returns Promise<CryptoKey> - The derived encryption key
 */
export async function deriveEncryptionKey(
	password: string,
	salt: string,
): Promise<CryptoKey> {
	const passwordBuffer = new TextEncoder().encode(password);
	const saltBuffer = hexToBytes(salt) as Uint8Array;

	// Import password as key material
	const keyMaterial = await crypto.subtle.importKey(
		"raw",
		passwordBuffer,
		"PBKDF2",
		false,
		["deriveKey"],
	);

	// Derive AES-GCM key using PBKDF2
	const key = await crypto.subtle.deriveKey(
		{
			name: "PBKDF2",
			salt: saltBuffer as BufferSource,
			iterations: PBKDF2_ITERATIONS,
			hash: "SHA-256",
		},
		keyMaterial,
		{ name: "AES-GCM", length: KEY_LENGTH },
		false, // not extractable (for security)
		["encrypt", "decrypt"],
	);

	return key;
}

/**
 * Encrypts a message using AES-GCM
 *
 * @param plaintext - The message to encrypt
 * @param key - The encryption key
 * @returns Promise<{ciphertext: string, nonce: string}> - Hex-encoded ciphertext and nonce
 */
export async function encryptMessage(
	plaintext: string,
	key: CryptoKey,
): Promise<{ ciphertext: string; nonce: string }> {
	// Generate random nonce
	const nonce = crypto.getRandomValues(new Uint8Array(NONCE_LENGTH));

	// Encrypt the message
	const plaintextBuffer = new TextEncoder().encode(plaintext);
	const ciphertextBuffer = await crypto.subtle.encrypt(
		{
			name: "AES-GCM",
			iv: nonce,
		},
		key,
		plaintextBuffer,
	);

	return {
		ciphertext: bytesToHex(new Uint8Array(ciphertextBuffer)),
		nonce: bytesToHex(nonce),
	};
}

/**
 * Decrypts a message using AES-GCM
 *
 * @param ciphertext - Hex-encoded ciphertext
 * @param nonce - Hex-encoded nonce
 * @param key - The decryption key
 * @returns Promise<string> - The decrypted plaintext
 * @throws Error if decryption fails (wrong key or corrupted data)
 */
export async function decryptMessage(
	ciphertext: string,
	nonce: string,
	key: CryptoKey,
): Promise<string> {
	try {
		const ciphertextBuffer = hexToBytes(ciphertext) as Uint8Array;
		const nonceBuffer = hexToBytes(nonce) as Uint8Array;

		const plaintextBuffer = await crypto.subtle.decrypt(
			{
				name: "AES-GCM",
				iv: nonceBuffer as BufferSource,
			},
			key,
			ciphertextBuffer as BufferSource,
		);

		return new TextDecoder().decode(plaintextBuffer);
	} catch (error) {
		throw new Error("Decryption failed: invalid key or corrupted message");
	}
}

/**
 * Generates a random hex-encoded salt for room encryption
 *
 * @returns string - Hex-encoded random salt
 */
export function generateSalt(): string {
	const salt = crypto.getRandomValues(new Uint8Array(SALT_LENGTH));
	return bytesToHex(salt);
}
