document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('start-button').addEventListener('click', () => {
        fetch('wolbolt.cgi/wol', {
            method: 'POST'
        })
        .then(response => {
            if (response.ok) {
                alert('起動したと思う');
            } else {
                console.error('failed to launch: %o', response)
                alert('起動失敗');
            }
        })
        .catch(error => {
            console.error('Error: %o', error);
            alert('起動失敗');
        });
    });
});
