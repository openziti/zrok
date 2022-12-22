const Detail = (props) => {
    return (
        <div className={"detail-container"}>
            <h1>{props.selection.id} ({props.selection.type})</h1>
        </div>
    );
};

export default Detail;