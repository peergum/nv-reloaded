'use client'

import Image from "next/image";
import NavBar from "@/components/NavBar";
import {useState} from "react";
import {menuLogo, menuOptions} from "@/app/menu";
import {BottomBar} from "@/components/BottomBar";

export default function Files() {
    return (
        <div className="flex flex-col h-screen border border-black">
            <NavBar menu={1}/>
            <main className="h-full mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 w-full">
                <div>files</div>
            </main>
            <BottomBar/>
        </div>
    );
}
