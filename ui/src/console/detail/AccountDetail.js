const AccountDetail = (props) => {
    return (
        <div>
            <h2>Your Account: {props.user.email}</h2>
        </div>
    );
}

export default AccountDetail;