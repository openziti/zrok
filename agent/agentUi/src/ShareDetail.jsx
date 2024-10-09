import {useParams} from "react-router-dom";

const ShareDetail = (props) => {
    let params = useParams();

    return (
        <>
            <h1>Share {params.token}</h1>
        </>
    )
}

export default ShareDetail;