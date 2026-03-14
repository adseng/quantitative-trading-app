import { NavLink, Outlet } from 'react-router-dom'
import { Suspense } from 'react'
import { twMerge } from 'tailwind-merge'
import HeaderMenu from './HeaderMenu'
import { MENUS } from './Menus'

const HEADER_HEIGHT = 50

export default function Layout() {
  return (
    <div className={twMerge('h-screen w-screen bg-white flex flex-col')}>
      <header
        className={twMerge(
          'fixed top-0 left-0 right-0 border-b border-gray-200 px-4 bg-white flex items-center justify-between'
        )}
        style={{ height: `${HEADER_HEIGHT}px` }}
      >
        <nav className="flex gap-4">
          {MENUS.map((menu) => (
            <NavLink
              key={menu.value}
              to={menu.value}
              style={({ isActive }) => ({
                color: isActive ? '#ec4899' : '#9ca3af',
                fontWeight: isActive ? 500 : 400,
              })}
              className={({ isActive }) =>
                twMerge(
                  'flex items-center gap-1 no-underline transition-colors',
                  isActive ? '' : 'hover:text-gray-500'
                )
              }
            >
              {menu.icon}
              {menu.label}
            </NavLink>
          ))}
        </nav>
        <HeaderMenu />
      </header>
      <main
        className="w-full flex-1 p-4 overflow-auto"
        style={{ marginTop: `${HEADER_HEIGHT}px` }}
      >
        <Suspense fallback={<div className="p-4">加载中...</div>}>
          <Outlet />
        </Suspense>
      </main>
    </div>
  )
}
