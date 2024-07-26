import {MenuItem} from "@/app/menu";

export function MenuLogo({logo}: { logo: MenuItem }) {
    return (
        <img
            alt={logo.name}
            src={logo.link}
            className="h-20 w-auto"
        />
    );
}
