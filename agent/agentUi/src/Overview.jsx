import "bootstrap/dist/css/bootstrap.min.css";
import ShareCard from "./ShareCard.jsx";
import AccessCard from "./AccessCard.jsx";

const Overview = (props) => {
    let cards = [];
    props.overview.forEach((row) => {
        switch(row.type) {
            case "share":
                cards.push(<ShareCard share={row.v} />);
                break;

            case "access":
                cards.push(<AccessCard access={row.v} />);
                break;
        }

    });

    return (
        <>
            {cards}
        </>
    )
}

export default Overview;
