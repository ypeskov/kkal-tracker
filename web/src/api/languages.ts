export interface Language {
    code: string;
    name: string;
}

const API_BASE = '/api';

export const languagesService = {
    getLanguages: async (): Promise<Language[]> => {
        const response = await fetch(`${API_BASE}/languages`);
        if (!response.ok) {
            throw new Error('Failed to fetch languages');
        }
        return response.json();
    },
};

