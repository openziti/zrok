import SecretToggle from "../../SecretToggle";
import PropertyTable from "../../PropertyTable";

const DetailTab = (props) => {
    const customProperties = {
        zId: row => <SecretToggle secret={row.value} />
    }

    return (
        <PropertyTable object={props.frontend} custom={customProperties} />
    );
};

export default DetailTab;