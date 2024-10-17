//
//  wrapsApp.swift
//  wraps
//
//  Created by Arjun Krishnan on 10/17/24.
//

import SwiftUI

@main
struct wrapsApp: App {
    let persistenceController = PersistenceController.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(\.managedObjectContext, persistenceController.container.viewContext)
        }
    }
}
