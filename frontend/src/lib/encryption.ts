import CryptoJS from 'crypto-js';

// Encryption configuration
const PBKDF2_ITERATIONS = 100000;
const KEY_SIZE = 256; // bits
const SALT_SIZE = 128; // bits

/**
 * Derive an encryption key from a password and salt
 * @param password - User's password
 * @param salt - Salt for key derivation (hex string)
 * @returns Derived encryption key
 */
export function deriveKey(password: string, salt: string): string {
	const saltBytes = CryptoJS.enc.Hex.parse(salt);
	const key = CryptoJS.PBKDF2(password, saltBytes, {
		keySize: KEY_SIZE / 32, // CryptoJS uses 32-bit words
		iterations: PBKDF2_ITERATIONS
	});
	return key.toString();
}

/**
 * Generate a random salt for key derivation
 * @returns Random salt as hex string
 */
export function generateSalt(): string {
	return CryptoJS.lib.WordArray.random(SALT_SIZE / 8).toString();
}

/**
 * Encrypt content using AES-256-GCM
 * @param content - Plain text content to encrypt
 * @param key - Encryption key
 * @returns Encrypted content with IV (format: IV:EncryptedData)
 */
export function encrypt(content: string, key: string): string {
	const iv = CryptoJS.lib.WordArray.random(128 / 8); // 12 bytes for GCM
	const encrypted = CryptoJS.AES.encrypt(content, CryptoJS.enc.Hex.parse(key), {
		iv: iv,
		mode: CryptoJS.mode.GCM,
		padding: CryptoJS.pad.Pkcs7
	});
	return iv.toString() + ':' + encrypted.toString();
}

/**
 * Decrypt content using AES-256-GCM
 * @param encryptedData - Encrypted data with IV (format: IV:EncryptedData)
 * @param key - Encryption key
 * @returns Decrypted plain text
 */
export function decrypt(encryptedData: string, key: string): string {
	const parts = encryptedData.split(':');
	if (parts.length !== 2) {
		throw new Error('Invalid encrypted data format');
	}
	const iv = CryptoJS.enc.Hex.parse(parts[0]);
	const encrypted = parts[1];
	const decrypted = CryptoJS.AES.decrypt(
		encrypted,
		CryptoJS.enc.Hex.parse(key),
		{
			iv: iv,
			mode: CryptoJS.mode.GCM,
			padding: CryptoJS.pad.Pkcs7
		}
	);
	return decrypted.toString(CryptoJS.enc.Utf8);
}

/**
 * Generate a hash of the encryption key for server-side verification
 * The actual key is never sent to the server, only this hash
 * @param key - Encryption key
 * @returns SHA-256 hash of the key
 */
export function hashKey(key: string): string {
	return CryptoJS.SHA256(key).toString();
}

/**
 * Generate a content hash for integrity verification
 * @param content - Plain text content
 * @returns SHA-256 hash of the content
 */
export function hashContent(content: string): string {
	return CryptoJS.SHA256(content).toString();
}

/**
 * Initialize encryption: generate salt, derive key, and return key hash
 * Call this when user sets up their encryption password
 * @param password - User's password
 * @returns Object with salt, derived key, and key hash
 */
export function initializeEncryption(password: string) {
	const salt = generateSalt();
	const key = deriveKey(password, salt);
	const keyHash = hashKey(key);
	return {
		salt,
		key,
		keyHash
	};
}

/**
 * Count words in a string
 * @param text - Text to count words in
 * @returns Word count
 */
export function countWords(text: string): number {
	return text.trim().split(/\s+/).filter(word => word.length > 0).length;
}
