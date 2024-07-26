import Image from "next/image";
import NavBar from "@/components/NavBar";
import {BottomBar} from "@/components/BottomBar";

export default function About() {
    return (
        <div className="flex flex-col h-screen">
            <NavBar menu={2}/>
            <main className="flex flex-col items-center justify-center h-full mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 w-full py-4">
                <div className="flex flex-col bg-white items-center justify-center min-w-80 min-h-80 shadow-sm shadow-black ">
                    <div>test</div>
                    <div>test2</div>
                    <div>test3</div>
                </div>
            </main>
            <BottomBar/>
        </div>
    );
}
