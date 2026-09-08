interface CacheEntry<T> {
	value: T;
	expires?: number;
}

export class SimpleCache<K, V> {
	private cache = new Map<K, CacheEntry<V>>();
	private maxSize?: number;

	constructor(maxSize?: number) {
		this.maxSize = maxSize;
	}

	get(key: K): V | undefined {
		const entry = this.cache.get(key);

		if (!entry) return undefined;

		// Check if expired
		if (entry.expires && Date.now() > entry.expires) {
			this.cache.delete(key);
			return undefined;
		}

		return entry.value;
	}

	set(key: K, value: V, ttlMs?: number): void {
		// Remove oldest entry if at max size
		if (this.maxSize && this.cache.size >= this.maxSize && !this.cache.has(key)) {
			const firstKey = this.cache.keys().next().value;
			if (firstKey !== undefined) {
				this.cache.delete(firstKey);
			}
		}

		const entry: CacheEntry<V> = {
			value,
			expires: ttlMs ? Date.now() + ttlMs : undefined
		};

		this.cache.set(key, entry);
	}

	has(key: K): boolean {
		return this.get(key) !== undefined;
	}

	delete(key: K): boolean {
		return this.cache.delete(key);
	}

	clear(): void {
		this.cache.clear();
	}

	size(): number {
		// Clean expired entries and return actual size
		this.cleanExpired();
		return this.cache.size;
	}

	private cleanExpired(): void {
		const now = Date.now();
		for (const [key, entry] of this.cache.entries()) {
			if (entry.expires && now > entry.expires) {
				this.cache.delete(key);
			}
		}
	}
}

// OAuth2 data structure for caching
export interface OAuth2Data {
	state: string;
	codeVerifier: string;
	redirectUri?: string;
	timestamp: number;
}

export interface OTPData {
	email: string;
	code: string;
	flow: 'signin' | 'signup';
	timestamp: number;
}

export const oauthCache = new SimpleCache<string, OAuth2Data>(1000);
export const otpCache = new SimpleCache<string, OTPData>(1000);
