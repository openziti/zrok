import NavBar from "./NavBar.jsx";
import {useParams} from "react-router-dom";

const ShareDetail = (props) => {
    let params = useParams();

    return (
        <>
            <NavBar version={props.version} />

            <h1>Share {params.token}</h1>
        </>
    )
}

export default ShareDetail;