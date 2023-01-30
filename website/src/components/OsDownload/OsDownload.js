import React from "react";

function getOs() {
    const os = ['Windows', 'Linux', 'Mac']; // add your OS values
    return os.find(v=>navigator.appVersion.indexOf(v) >= 0);
}

export default function OsDownload() {
    return <p>i did it</p>;
}