import "bootstrap/dist/css/bootstrap.min.css";
import DataTable from "react-data-table-component";
import NavBar from "./NavBar.jsx";

const Overview = (props) => {
    const shareColumns = [
        {
            name: 'Token',
            selector: row => <a href={"/share/"+row.token}>{row.token}</a>
        },
        {
            name: 'Share Mode',
            selector: row => row.shareMode
        },
        {
            name: 'Backend Mode',
            selector: row => row.backendMode
        },
        {
            name: 'Target',
            selector: row => row.backendEndpoint,
        },
        {
            name: 'Frontend Endpoints',
            selector: row => <div>{row.shareMode === "public" ? row.frontendEndpoint.map((fe) => <a href={fe.toString()} target={"_"}>{fe}</a>) : "---"}</div>,
            grow: 2
        }
    ];

    const accessColumns = [
        {
            name: 'Frontend Token',
            selector: row => row.frontendToken
        },
        {
            name: 'Token',
            selector: row => row.token
        },
        {
            name: 'Bind Address',
            selector: row => row.bindAddress
        },
    ];

    return (
        <>
            <NavBar version={props.version} />

            <div className={"info"}>
                <h2>Shares</h2>
                <DataTable
                    columns={shareColumns}
                    data={props.shares}
                    noDataComponent={<div/>}
                />
            </div>

            <div className={"info"}>
                <h2>Accesses</h2>
                <DataTable
                    columns={accessColumns}
                    data={props.accesses}
                    noDataComponent={<div/>}
                />
            </div>
        </>
    )
}

export default Overview;
