import {useEffect, useState} from "react";
import {MetadataApi} from "./api";

const ApiConsole = () => {
    const [version, setVersion] = useState("no version set");

    useEffect(() => {
        let api = new MetadataApi();
        api.version()
            .then(d => {
                setVersion(d);
            })
            .catch(e => {
                console.log(e);
            });
    }, []);

    return (
        <div>
            <h1>zrok</h1>
            <h2>{version}</h2>
        </div>
    );
}

export default ApiConsole;