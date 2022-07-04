import App from './App.svelte';

const app = new App({ target: document.body,
	props: {
    notes: JSON.parse(window.notes) || [],
    previousDay: window.previousDay,
    nextDay: window.nextDay,
	}
});

export default app;
