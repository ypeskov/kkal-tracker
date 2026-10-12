import { filenameFromContentDisposition } from '@/utils/download';

export const FOOD_LOG_SKILL_FALLBACK_FILENAME = 'kkal-tracker-food-log.zip';

export interface SkillDownload {
  blob: Blob;
  filename: string;
}

// Thrown on a non-OK or unexpected response; keeps the status, never the server's response text
export class SkillDownloadError extends Error {
  readonly status: number;

  constructor(status: number) {
    super('Failed to download the skill');
    this.name = 'SkillDownloadError';
    this.status = status;
  }
}

class SkillsService {
  private getHeaders() {
    const token = sessionStorage.getItem('token');
    return {
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  }

  // Every call creates a new API key on the server and returns the skill zip with that key inside
  downloadFoodLogSkill = async (): Promise<SkillDownload> => {
    const response = await fetch('/api/skills/food-log/download', {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new SkillDownloadError(response.status);
    }

    // Guard against saving a non-zip body (e.g. the SPA fallback page) as the skill
    if (!response.headers.get('Content-Type')?.includes('application/zip')) {
      throw new SkillDownloadError(response.status);
    }

    const filename =
      filenameFromContentDisposition(response.headers.get('Content-Disposition')) ??
      FOOD_LOG_SKILL_FALLBACK_FILENAME;
    const blob = await response.blob();
    return { blob, filename };
  };
}

export const skillsService = new SkillsService();
