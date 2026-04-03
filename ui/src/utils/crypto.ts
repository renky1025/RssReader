/**
 * RSA encryption utilities using Web Crypto API
 */

export interface PublicKeyInfo {
  key_id: string
  public_key: string
  expires_at: number
}

/**
 * Encrypt password using RSA-OAEP with SHA-256
 */
export async function encryptPassword(password: string, publicKeyPEM: string): Promise<string> {
  // Remove PEM headers and decode base64
  const pemContents = publicKeyPEM
    .replace(/-----BEGIN PUBLIC KEY-----/g, '')
    .replace(/-----END PUBLIC KEY-----/g, '')
    .replace(/\s/g, '')
  
  const binaryDer = Uint8Array.from(atob(pemContents), c => c.charCodeAt(0))

  // Import the public key
  const publicKey = await crypto.subtle.importKey(
    'spki',
    binaryDer,
    {
      name: 'RSA-OAEP',
      hash: 'SHA-256',
    },
    false,
    ['encrypt']
  )

  // Encrypt the password
  const encrypted = await crypto.subtle.encrypt(
    { name: 'RSA-OAEP' },
    publicKey,
    new TextEncoder().encode(password)
  )

  // Return base64 encoded ciphertext
  return btoa(String.fromCharCode(...new Uint8Array(encrypted)))
}

/**
 * Generate a random nonce for replay protection
 */
export function generateNonce(): string {
  const array = new Uint8Array(16)
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * Get current timestamp in seconds
 */
export function getTimestamp(): number {
  return Math.floor(Date.now() / 1000)
}
