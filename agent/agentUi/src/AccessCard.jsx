import LanIcon from "@mui/icons-material/Lan";

const AccessCard = (props) => {
    return (
        <div className={"card"}>
            <h2>{props.access.frontendToken} [<LanIcon />]</h2>
            <p>
                {props.access.token} &rarr; {props.access.bindAddress}
            </p>
        </div>
    );
}

export default AccessCard;