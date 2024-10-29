export const buildOverview = (status) => {
    let overview = [];
    if(status) {
        if(status.accesses) {
            status.accesses.forEach(acc => {
                let o = structuredClone(acc);
                o["type"] = "access";
                o["id"] = acc.frontendToken;
                overview.push(o);
            });
        }
        if(status.shares) {
            status.shares.forEach(shr => {
                let o = structuredClone(shr);
                o["type"] = "share";
                o["id"] = shr.token;
                overview.push(o);
            });
        }
    }
    overview.sort((a, b) => {
        if(a.id < b.id) return -1;
        if(a.id > b.id) return 1;
    });
    return overview;
}