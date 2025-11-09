import { render, screen } from '@testing-library/react';
import user from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Login from './Login';

test('renders login form and validates submit', async () => {
  render(<MemoryRouter><Login /></MemoryRouter>);
  expect(screen.getByText(/Welcome back/i)).toBeInTheDocument();

  // mock fetch
  global.fetch = vi.fn().mockResolvedValue({
    ok: true, status: 200, json: async () => ({ token: 't' }),
  });

  await user.type(screen.getByPlaceholderText(/Email/), 'a@b.com');
  await user.type(screen.getByPlaceholderText(/Password/), 'secret123');
  await user.click(screen.getByRole('button', { name: /login/i }));

  expect(global.fetch).toHaveBeenCalled();
});
