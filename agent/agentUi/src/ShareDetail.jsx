import NavBar from "./NavBar.jsx";
import {useParams} from "react-router-dom";

function ShareDetail() {
    let params = useParams();

    return (
        <>
            <NavBar />

            <h1>Share {params.token}</h1>
        </>
    )
}

export default ShareDetail;