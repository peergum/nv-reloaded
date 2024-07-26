'use client'

import Image from "next/image";
import NavBar from "@/components/NavBar";
import {useState} from "react";
import {menuLogo, menuOptions} from "@/app/menu";
import {setActiveMenu} from "@/app/layout";
import {BottomBar} from "@/components/BottomBar";
import Form from "@/components/Form";

const generalSettings = "text-lg border-b-2"
export default function Setup() {
    return (
        <div className="flex flex-col border border-black">
            <NavBar menu={0}/>
            <main className="mt-16 m-4 bg-white shadow-sm shadow-black border border-gray-500 mx-auto max-w-7xl w-full">
                <Form/>
            </main>
            <BottomBar/>
        </div>
    );
}
