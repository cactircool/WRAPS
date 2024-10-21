//
//  ContentView.swift
//  wraps
//
//  Created by Arjun Krishnan on 10/17/24.
//

import SwiftUI
import CoreData

struct ContentView: View {
    var body: some View {
        TabView {
            Tab("", systemImage: "house") {
                HomeView()
            }
            
            Tab("", systemImage: "magnifyingglass") {
                SearchView()
            }
            
            Tab("", systemImage: "doc") {
                DocumentsView()
            }
            
            Tab("", systemImage: "gearshape") {
                SettingsView()
            }
        }
    }
}

#Preview {
    ContentView().environment(\.managedObjectContext, PersistenceController.preview.container.viewContext)
}
