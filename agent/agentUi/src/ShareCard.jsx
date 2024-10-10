import ShareIcon from "@mui/icons-material/Share";
import DeleteIcon from "@mui/icons-material/Delete";

const ShareCard = (props) => {
    let frontends = [];
    props.share.frontendEndpoint.map((fe) => {
        frontends.push(<a href={fe.toString()} target={"_"}>{fe}</a>);
    })

    const releaseClicked = () => {
        props.releaseShare({token: props.share.token}, (err, data) => { console.log("releaseClicked", data); });
    }

    return (
        <div className={"card"}>
            <h2>{props.share.token} [<ShareIcon />]</h2>
            <p>({props.share.shareMode}, {props.share.backendMode})</p>
            <p>
                {props.share.backendEndpoint} &rarr; {frontends}
            </p>
            <p><DeleteIcon onClick={releaseClicked}/></p>
        </div>
    );
}

export default ShareCard;