import DataTable from 'react-data-table-component';
import Services from './Services';

const Environments = (props) => {
    const humanizeDuration = require("humanize-duration")

    const columns = [
        {
            name: 'Description',
            selector: row => row.environment.description,
            sortable: true,
        },
        {
            name: 'Host',
            selector: row => row.environment.host,
            sortable: true,
        },
        {
            name: 'Address',
            selector: row => row.environment.address,
            sortable: true,
        },
        {
            name: 'Identity',
            selector: row => row.environment.zId,
            sortable: true,
        },
        {
            name: 'Uptime',
            selector: row => humanizeDuration(new Date().getTime() - row.environment.updatedAt),
            sortable: true,
        },
    ]

    const servicesComponent = ({ data }) => <Services services={data.services} />
    const servicesExpanded = row => row.services != null && row.services.length > 0

    console.log('now', Date.now())

    return (
        <div>
            <h2>Environments</h2>
            { props.overview && props.overview.length > 0 && (
                <div>
                    <DataTable
                        columns={columns}
                        data={props.overview}
                        defaultSortFieldId={1}
                        expandableRows
                        expandableRowsComponent={servicesComponent}
                        expandableRowExpanded={servicesExpanded}
                    />
                </div>
            )}
        </div>
    )
};

export default Environments;
