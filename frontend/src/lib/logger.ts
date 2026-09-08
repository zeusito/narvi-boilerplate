import pino from 'pino';
import { dev } from '$app/environment';

export const logger = pino({
	level: dev ? 'debug' : 'info',
	formatters: {
		level: (label) => {
			return {
				level: label
			};
		}
	},
	timestamp: pino.stdTimeFunctions.isoTime
});
