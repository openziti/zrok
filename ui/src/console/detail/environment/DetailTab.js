import PropertyTable from "../../PropertyTable";
import SecretToggle from "../../SecretToggle";

const DetailTab = (props) => {
    const customProperties = {
        zId: row => <SecretToggle secret={row.value} />
    }

    return (
        <PropertyTable object={props.environment} custom={customProperties} />
    );
};

export default DetailTab;