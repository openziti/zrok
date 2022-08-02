import {useEffect, useState} from "react";
import * as metadata from "./api/metadata";

const Version = () => {
    const [v, setV] = useState('');

    useEffect(() => {
        let mounted = true;
        metadata.version().then(resp => {
            if(mounted) {
                setV(resp.data);
            }
        });
        return () => {
            mounted = false;
        };
    }, []);

    return (
        <h2>{v}</h2>
    );
}

export default Version;