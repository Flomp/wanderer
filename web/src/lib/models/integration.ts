
interface BaseIntegration {
    active: boolean
}

interface StravaIntegration extends BaseIntegration {
    clientId: string | number;
    clientSecret: string;
    routes: boolean;
    activities: boolean;
    accessToken?: string;
    refreshToken?: string;
    expiresAt?: number;
}

export class Integration {
    id?: string;
    user: string;
    strava?: StravaIntegration;

    constructor(user: string, strava?: StravaIntegration) {
        this.user = user;
        this.strava = strava;
    }
}