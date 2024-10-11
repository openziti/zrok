import LanIcon from "@mui/icons-material/Lan";
import DeleteIcon from "@mui/icons-material/Delete";

const AccessCard = (props) => {
    const releaseClicked = () => {
        props.releaseAccess({frontendToken: props.access.frontendToken}, (err, data) => { console.log("releaseClicked", data); });
    }

    return (
        <div className={"card"}>
            <h2>{props.access.frontendToken} [<LanIcon/>]</h2>
            <p>
                {props.access.token} &rarr; {props.access.bindAddress}
            </p>
            <p><DeleteIcon onClick={releaseClicked}/></p>
        </div>
    );
}

export default AccessCard;