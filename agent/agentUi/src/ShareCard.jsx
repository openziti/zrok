import ShareIcon from "@mui/icons-material/Share";

const ShareCard = (props) => {
    let frontends = [];
    props.share.frontendEndpoint.map((fe) => {
        frontends.push(<a href={fe.toString()} target={"_"}>{fe}</a>);
    })

    return (
        <div className={"card"}>
            <h2>{props.share.token} [<ShareIcon />]</h2>
            <p>({props.share.shareMode}, {props.share.backendMode})</p>
            <p>
                {props.share.backendEndpoint} &rarr; {frontends}
            </p>
        </div>
    );
}

export default ShareCard;