import {useEffect, useState} from "react";
import DataTable from "react-data-table-component";

const objectToRows = (obj) => {
    let rows = [];
    for(const key in obj) {
        rows.push({
            property: key,
            value: obj[key]
        });
    }
    return rows;
};

const camelToWords = (s) => s.replace(/([A-Z])/g, ' $1').replace(/^./, function(str){ return str.toUpperCase(); });

const rowToValue = (row) => {
    if(row.property.endsWith("At")) {
        return new Date(row.value).toLocaleString();
    }
    return row.value;
};

const PropertyTable = (props) => {
    const [data, setData] = useState([]);

    useEffect(() => {
        setData(objectToRows(props.object));
    }, [props.object]);

    const columns = [
        {
            name: "Property",
            selector: row => camelToWords(row.property),
            sortable: true
        },
        {
            name: "Value",
            selector: row => rowToValue(row),
            sortable: true,
            grow: 3
        }
    ];

    return <DataTable columns={columns} data={data} className={"zrok-datatable"} />;
};

export default PropertyTable;