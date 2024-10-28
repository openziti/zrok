export const buildOverview = (status) => {
    let overview = [];
    if(status) {
        if(status.accesses) {
            status.accesses.forEach(acc => {
                let o = structuredClone(acc);
                o["type"] = "access";
                overview.push(o);
            });
        }
        if(status.shares) {
            status.shares.forEach(shr => {
                let o = structuredClone(shr);
                o["type"] = "share";
                overview.push(o);
            });
        }
    }
    return overview;
}