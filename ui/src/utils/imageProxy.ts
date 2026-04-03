/**
 * Image proxy utility to solve 403 errors from RSS images with hotlink protection.
 * This module provides functions to convert external image URLs to use the local proxy.
 */

/**
 * Convert an image URL to use the image proxy.
 * Only converts http/https URLs, leaves other URLs unchanged.
 */
export function getProxyImageUrl(url: string): string {
    if (!url) return url

    // Only proxy http/https URLs
    if (!url.startsWith('http://') && !url.startsWith('https://')) {
        return url
    }

    // Don't proxy data URLs or already proxied URLs
    if (url.startsWith('data:') || url.includes('/api/v1/proxy')) {
        return url
    }

    return `/api/v1/proxy?url=${encodeURIComponent(url)}`
}

/**
 * Process HTML content to replace all image URLs with proxied versions.
 * This handles images in article content that use hotlink protection.
 */
export function processContentImages(html: string): string {
    if (!html) return html

    // Match img tags with src attribute
    // Handles both single and double quotes, and various attribute orderings
    return html.replace(
        /<img([^>]*?)src=["']([^"']+)["']([^>]*?)>/gi,
        (_match, before, src, after) => {
            const proxiedSrc = getProxyImageUrl(src)
            return `<img${before}src="${proxiedSrc}"${after}>`
        }
    )
}
