export interface MenuItem {
    name: string,
    link: string,
}

export const menuOptions: MenuItem[] = [
    {name: 'Setup', link: '/'},
    {name: 'Files', link: '/files'},
    {name: 'About', link: '/about'},
]

export const menuLogo: MenuItem = {
    name: "NV-Reloaded",
    link: "/nv-reloaded-logo.png"
}

