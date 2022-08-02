import {useEffect} from "react";
import * as metadata from './api/metadata';

const Identities = (props) => {
    useEffect(() => {
        metadata.listIdentities().then((resp) => { console.log(resp); })
    }, [])

    return (
        <div>
            <h3>Identities</h3>
        </div>
    )
};

export default Identities;
