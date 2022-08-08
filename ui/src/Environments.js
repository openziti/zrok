import DataTable from 'react-data-table-component';
import Services from './Services';

const Environments = (props) => {
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
            selector: row => row.environment.zitiIdentityId,
            sortable: true,
        },
        {
            name: 'Updated',
            selector: row => row.environment.updatedAt,
            sortable: true,
        },
    ]

    const conditionalRowStyles = [
        {
            when: row => !row.environment.active,
            style: {
                display: 'none'
            }
        }
    ]

    const servicesComponent = ({ data }) => <Services services={data.services} />
    const servicesExpanded = row => row.services != null && row.services.length > 0 && row.services.some((row) => row.active)

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
                        conditionalRowStyles={conditionalRowStyles}
                    />
                </div>
            )}
        </div>
    )
};

export default Environments;
