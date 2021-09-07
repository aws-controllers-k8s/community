window.addEventListener('DOMContentLoaded', (event) => {
  document.getElementById('mode').addEventListener('click', () => {
    document.body.classList.toggle('dark');
    localStorage.setItem('theme', document.body.classList.contains('dark') ? 'dark' : 'light');
  });

  let systemPreferenceDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  let savedPreference = localStorage.getItem('theme');
  if (savedPreference === 'light' || (savedPreference === null && !systemPreferenceDark)) {
    document.body.classList.remove('dark');
  }
});
