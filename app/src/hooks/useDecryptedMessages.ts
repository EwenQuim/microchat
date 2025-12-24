import { useEffect, useState } from "react";
import type { Message } from "@/lib/api/generated/openAPI.schemas";
import { decryptMessage } from "@/lib/crypto/e2e";

/**
 * Hook that decrypts encrypted messages
 *
 * @param messages - Array of messages (potentially encrypted)
 * @param encryptionKey - Encryption key for the room (if available)
 * @returns Array of messages with decrypted content
 */
export function useDecryptedMessages(
	messages: Message[] | undefined,
	encryptionKey: CryptoKey | undefined,
): Message[] {
	const [decryptedMessages, setDecryptedMessages] = useState<Message[]>([]);

	useEffect(() => {
		if (!messages) {
			setDecryptedMessages([]);
			return;
		}

		const decryptMessages = async () => {
			const results = await Promise.all(
				messages.map(async (message) => {
					// If message is not encrypted, return as-is
					if (!message.is_encrypted || !message.nonce) {
						return message;
					}

					// If message is encrypted but no key available, show placeholder
					if (!encryptionKey) {
						return {
							...message,
							content: "[Encrypted message - password required]",
						};
					}

					// Decrypt the message
					try {
						const decrypted = await decryptMessage(
							message.content || "",
							message.nonce || "",
							encryptionKey,
						);
						return {
							...message,
							content: decrypted,
						};
					} catch (error) {
						console.error("Failed to decrypt message:", error);
						return {
							...message,
							content: "[Failed to decrypt message]",
						};
					}
				}),
			);

			setDecryptedMessages(results);
		};

		decryptMessages();
	}, [messages, encryptionKey]);

	return decryptedMessages;
}
