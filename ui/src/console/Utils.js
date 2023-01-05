export const secretString = (s) => {
    let out = "";
    for(let i = 0; i < s.length; i++) {
        out += "*";
    }
    return out;
}