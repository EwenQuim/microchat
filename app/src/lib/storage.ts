const USERNAME_KEY = "microchat_username";

export const storage = {
	getUsername: (): string | null => {
		return localStorage.getItem(USERNAME_KEY);
	},

	setUsername: (username: string): void => {
		localStorage.setItem(USERNAME_KEY, username);
	},

	clearUsername: (): void => {
		localStorage.removeItem(USERNAME_KEY);
	},
};
