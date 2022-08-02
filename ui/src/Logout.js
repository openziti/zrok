const Logout = (props) => {
    const onClick = () => {
        props.logout()
    }

    return (
        <a onClick={onClick}>[x] {props.user.email}</a>
    );
}

export default Logout;
