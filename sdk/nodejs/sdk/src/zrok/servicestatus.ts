
export enum Permissions {
    None = 0,
    Bind = 2,
    Dial = 3,
}


export class ServiceStatus {
    private status: number | undefined;
    private permissions: Permissions | undefined;

    constructor(statusInfo: { status: number; permissions: Permissions }) {
        this.status = statusInfo.status;
        this.permissions = statusInfo.permissions;
    }

    public getStatus(): number | undefined {
        return this.status;
    }

    public getPermissions(): Permissions | undefined {
        return this.permissions;
    }
}

