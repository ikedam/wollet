document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('start-button').addEventListener('click', () => {
        fetch('wolbolt.cgi/wol', {
            method: 'POST'
        })
        .then(response => {
            if (response.ok) {
                alert('WOL command sent successfully!');
            } else {
                alert('Failed to send WOL command.');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('An error occurred.');
        });
    });
});